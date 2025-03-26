package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/newrelic/node-log-viewer/internal/common"
	"github.com/newrelic/node-log-viewer/internal/database"
	log "github.com/newrelic/node-log-viewer/internal/log"
	"github.com/newrelic/node-log-viewer/internal/misc"
	"github.com/newrelic/node-log-viewer/internal/tui"
	v0 "github.com/newrelic/node-log-viewer/internal/v0"
	"github.com/spf13/afero"
	flag "github.com/spf13/pflag"
	"io"
	"log/slog"
	"os"
	"path"
	"regexp"
	"strings"
)

var exitStatus int
var fs = afero.NewOsFs()
var logger *log.Logger
var db *database.LogsDatabase

func main() {
	exitStatus = 0
	err := run(os.Args)
	if err != nil {
		fmt.Printf("app error: %v\n", err)
		exitStatus = 1
	}
	os.Exit(exitStatus)
}

func run(args []string) error {
	err := createAndParseFlags(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if flags.Version == true {
		misc.Version()
		return nil
	}

	logLevel := slog.LevelInfo
	if flags.LogLevel.String() != "" {
		logLevel = flags.LogLevel.ToLeveler().Level()
	}
	logger, _ = log.New(log.WithLevel(logLevel))
	logger.Debug("app flags", "flags", flags.String())

	db, err = initializeDatabase(logger)
	logger.Info("cache file created", "cache-file", db.DatabaseFile)
	defer shutdownDatabase(db, logger)

	// TODO: load input file in a gofunc so we can render a progress bar
	selectResult, err := db.GetAllLogs()
	if err != nil {
		logger.Error("could not verify cache", "error", err)
		return err
	}

	var lines []common.Envelope
	if len(selectResult.Rows) > 0 {
		logger.Debug("restored lines from cache file")
		lines = selectResult.ToLines()
	} else {
		var inputFile io.Reader

		switch {
		case flags.InputFile != "":
			logger.Debug("attempting to parse log file (-f)", "log-file", flags.InputFile)
			inputFile, err = openLogFile(flags.InputFile, logger)

		case len(flags.PositionalArgs) == 1:
			logger.Debug("attempting to parse log file (positional)", "log-file", flags.PositionalArgs[0])
			inputFile, err = openLogFile(flags.PositionalArgs[0], logger)
		}

		if err != nil {
			logger.Error("could not open log file", "error", err)
			return err
		}

		lines, err = parseLogFile(inputFile, db, logger)
		if err != nil {
			logger.Debug("could not parse log file", "error", err)
			return err
		}
	}

	if flags.DumpRemotePayloads == true {
		logger.Debug("dumping remote payloads")
		err = dumpRemotePayloads(db, os.Stdout)
		if err != nil {
			logger.Error("error dumping remote payloads", "error", err)
		}
		return nil
	}

	logger.Debug("starting tui")
	ui := tui.NewTUI(lines, db, logger)
	err = ui.App.Run()
	if err != nil {
		logger.Error("tui application error", "error", err)
		return err
	}

	return nil
}

// dbFile creates a temporary file and returns the path to it.
func dbFile() (string, error) {
	tmpDir := os.TempDir()
	return path.Join(tmpDir, "newrelic_agent.sqlite"), nil
}

func initializeDatabase(logger *log.Logger) (*database.LogsDatabase, error) {
	var databaseFile string
	if flags.CacheFile == "" {
		databaseFile, _ = dbFile()
	} else {
		databaseFile = flags.CacheFile
	}

	d, err := database.New(database.DbParams{
		DatabaseFilePath: databaseFile,
		Logger:           logger,
		DoMigration:      true,
	})
	if err != nil {
		return nil, err
	}

	return d, nil
}

func shutdownDatabase(db *database.LogsDatabase, logger *log.Logger) {
	db.Close()
	if flags.KeepCacheFile == true {
		return
	}
	err := os.Remove(db.DatabaseFile)
	if err != nil {
		logger.Error("failed to remove cache file", "cache-file", db.DatabaseFile, "error", err)
		exitStatus = 1
	}
}

func openLogFile(filePath string, logger *log.Logger) (io.Reader, error) {
	if filePath == "" {
		return nil, fmt.Errorf("no input log file provided")
	}

	logger.Debug("opening log file", "file", filePath)
	return fs.Open(filePath)
}

var matchLeadingK8sTimestamp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}\.\d{3,}\s+?`)

func parseLogFile(logFile io.Reader, db *database.LogsDatabase, logger *log.Logger) ([]common.Envelope, error) {
	lines := make([]common.Envelope, 0)
	scanner := bufio.NewScanner(logFile)

	for scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			logger.Error("failed to scan input", "error", err)
			return nil, err
		}

		// TODO: when we have a v1 line type, we need to do some text inspection
		// to determine the type of log line to unmarshal. This may mean we need
		// different viewers, as well.
		var envelope *v0.LineEnvelope
		sourceBytes := scanner.Bytes()
		sourceString := string(sourceBytes)

		if len(sourceString) == 0 {
			// Skip empty lines in the source file.
			continue
		}
		if matchLeadingK8sTimestamp.MatchString(sourceString) == true {
			// Looks like the line starts with a k8s style timestamp. So we
			// trim it off.
			idx := strings.Index(sourceString, "{")
			if idx == -1 {
				logger.Warn("line with leading timestamp does not seem to be a log line", "line", sourceString)
				continue
			}
			logger.Debug("trimming leading timestamp", "line", sourceString)
			sourceString = sourceString[idx:]
		}
		if sourceString[0:1] != "{" || sourceString[len(sourceString)-1:] != "}" {
			// Skip lines that do not look like NDJSON.
			logger.Warn("skipping parsing of malformed line", "line", sourceString)
			continue
		}

		err = json.Unmarshal([]byte(sourceString), &envelope)
		if err != nil {
			logger.Error("failed to parse line", "error", err, "line", sourceString)
			return nil, fmt.Errorf("%w: `%s`", err, sourceString)
		}

		err = db.Insert(envelope, sourceString)
		if err != nil {
			logger.Error("failed to insert line into database", "error", err, "line", sourceString)
			return nil, fmt.Errorf("%w: `%s`", err, sourceString)
		}

		lines = append(lines, envelope)
	}
	logger.Debug("finished reading log lines from input")

	return lines, nil
}

func dumpRemotePayloads(db *database.LogsDatabase, writer io.Writer) error {
	// TODO: if we implement a search by "component", utilize that here instead
	results, err := db.Search("remote_method")
	if err != nil {
		return fmt.Errorf("failed to search logs for remote payloads: %w", err)
	}

	for _, result := range results.Rows {
		io.WriteString(writer, result.Original+"\n")
	}

	return nil
}
