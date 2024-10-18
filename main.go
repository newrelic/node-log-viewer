package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jsumners-nr/nr-node-logviewer/internal/common"
	"github.com/jsumners-nr/nr-node-logviewer/internal/database"
	log "github.com/jsumners-nr/nr-node-logviewer/internal/log"
	"github.com/jsumners-nr/nr-node-logviewer/internal/tui"
	v0 "github.com/jsumners-nr/nr-node-logviewer/internal/v0"
	"github.com/spf13/afero"
	flag "github.com/spf13/pflag"
	"io"
	"log/slog"
	"os"
	"path"
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

	logLevel := slog.LevelInfo
	if flags.logLevel.String() != "" {
		logLevel = flags.logLevel.ToLeveler().Level()
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
	if len(selectResult.Lines) > 0 {
		logger.Debug("restored lines from cache file")
		lines = selectResult.Lines
	} else {
		logger.Debug("attempting to parse log file", "log-file", flags.inputFile)
		inputFile, err := openLogFile(flags.inputFile, logger)
		if err != nil {
			logger.Error("could not open log file", "error", err)
			return err
		}

		lines, err = parseLogFile(inputFile, logger)
		if err != nil {
			logger.Debug("could not parse log file", "error", err)
			return err
		}
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
	if flags.cacheFile == "" {
		databaseFile, _ = dbFile()
	} else {
		databaseFile = flags.cacheFile
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
	if flags.keepCacheFile == true {
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

func parseLogFile(logFile io.Reader, logger *log.Logger) ([]common.Envelope, error) {
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
		err = json.Unmarshal(sourceBytes, &envelope)
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
