package main

import (
	"github.com/jsumners-nr/nr-node-logviewer/internal/database"
	"github.com/jsumners-nr/nr-node-logviewer/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		assert.Equal(t, 2, len(lines))
		assert.Equal(t, "Wrapping 8 properties on nodule.", lines[0].Message())
		assert.Equal(t, `Replacing "all" with wrapped version`, lines[1].Message())
	})
}
