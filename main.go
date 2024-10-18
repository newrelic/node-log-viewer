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

	var databaseFile string
	if flags.cacheFile == "" {
		databaseFile, _ = dbFile()
	} else {
		databaseFile = flags.cacheFile
	}
	db, err = database.New(database.DbParams{
		DatabaseFilePath: databaseFile,
		Logger:           logger,
		DoMigration:      true,
	})
	if err != nil {
		return err
	}
	logger.Info("cache file created", "cache-file", databaseFile)
	defer db.Close()
	defer func() {
		if flags.keepCacheFile == true {
			return
		}
		e := os.Remove(databaseFile)
		if e != nil {
			logger.Error("failed to remove cache file", "cache-file", databaseFile, "error", e)
			exitStatus = 1
		}
	}()

	// TODO: load input file in a gofunc so we can render a progress bar
	var inputFile io.Reader
	if flags.inputFile == "" {
		logger.Error("no input file provided")
		return fmt.Errorf("no input file provided")
	} else {
		logger.Debug("reading log lines from file", "file", flags.inputFile)
		inputFile, err = fs.Open(flags.inputFile)
		if err != nil {
			return err
		}
	}

	// TODO: check if the cache already has logs. If so, skip parsing the file.
	lines := make([]common.Envelope, 0)
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		err = scanner.Err()
		if err != nil {
			logger.Error("failed to scan input", "error", err.Error())
			return err
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
			return fmt.Errorf("%w: `%s`", err, sourceString)
		}

		err = db.Insert(envelope, sourceString)
		if err != nil {
			logger.Error("failed to insert line into database", "error", err, "line", sourceString)
			return fmt.Errorf("%w: `%s`", err, sourceString)
		}

		lines = append(lines, envelope)
	}
	logger.Debug("finished reading log lines from input")

	logger.Debug("starting tui")
	ui := tui.NewTUI(lines, logger)
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
