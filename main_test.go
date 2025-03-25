package main

import (
	"github.com/newrelic/node-log-viewer/internal/database"
	"github.com/newrelic/node-log-viewer/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

var nullLogger = log.NewDiscardLogger()

func Test_parseLogFile(t *testing.T) {
	t.Run("handles node warnings embedded in file", func(t *testing.T) {
		testDb, err := database.New(database.DbParams{
			DatabaseFilePath: "file::memory:",
			DoMigration:      true,
			Logger:           nullLogger,
		})
		require.Nil(t, err)

		reader, err := fs.Open("testdata/node-warning-included.ndjson")
		require.Nil(t, err)

		lines, err := parseLogFile(reader, testDb, nullLogger)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(lines))
		assert.Equal(t, "Wrapping 8 properties on nodule.", lines[0].Message())
		assert.Equal(t, `Replacing "all" with wrapped version`, lines[1].Message())
	})

	t.Run("handles k8s-style prefixed lines", func(t *testing.T) {
		testDb, err := database.New(database.DbParams{
			DatabaseFilePath: "file::memory:",
			DoMigration:      true,
			Logger:           nullLogger,
		})
		require.Nil(t, err)

		reader, err := fs.Open("testdata/k8s-interleaved.log")
		require.Nil(t, err)

		lines, err := parseLogFile(reader, testDb, nullLogger)
		assert.Nil(t, err)
		assert.Equal(t, 8, len(lines))
		assert.Equal(t, "Created segment", lines[7].Message())
	})

	t.Run("dumps remote_method logs to stdout", func(t *testing.T) {
		testDb, err := database.New(database.DbParams{
			DatabaseFilePath: "file::memory:",
			DoMigration:      true,
			Logger:           nullLogger,
		})
		require.Nil(t, err)

		reader, err := fs.Open("testdata/v0/http-server.log")
		require.Nil(t, err)

		_, err = parseLogFile(reader, testDb, nullLogger)
		require.Nil(t, err)

		writer := &strings.Builder{}
		err = dumpRemotePayloads(testDb, writer)
		require.Nil(t, err)
		logs := strings.Split(writer.String(), "\n")
		assert.Equal(t, 384, len(logs))
	})
}
