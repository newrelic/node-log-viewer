package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jsumners-nr/nr-node-logviewer/internal/common"
	log "github.com/jsumners-nr/nr-node-logviewer/internal/log"
	"github.com/jsumners-nr/nr-node-logviewer/internal/tui"
	v0 "github.com/jsumners-nr/nr-node-logviewer/internal/v0"
	"github.com/spf13/afero"
	flag "github.com/spf13/pflag"
	"io"
	"log/slog"
	"os"
)

var fs = afero.NewOsFs()
var logger *log.Logger

func main() {
	err := run(os.Args)
	if err != nil {
		fmt.Printf("app error: %v\n", err)
		os.Exit(1)
	}
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
		err = json.Unmarshal(scanner.Bytes(), &envelope)
		if err != nil {
			offendingString := string(scanner.Bytes())
			logger.Error("failed to parse line", "error", err.Error(), "line", offendingString)
			return fmt.Errorf("%w: `%s`", err, offendingString)
		}

		lines = append(lines, envelope)
	}
	logger.Debug("finished reading log lines from input")

	logger.Debug("starting tui")
	appModel := tui.NewAppModel(lines, logger)
	app := tea.NewProgram(appModel, tea.WithAltScreen(), tea.WithMouseAllMotion())
	_, err = app.Run()
	if err != nil {
		logger.Error("tui application error", "error", err.Error())
		return err
	}

	return nil
}
