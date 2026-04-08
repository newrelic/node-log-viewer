package database

import (
	"testing"

	"github.com/newrelic/node-log-viewer/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var nullLogger = log.NewDiscardLogger()

func TestQuery(t *testing.T) {
	testDb, err := New(DbParams{
		DatabaseFilePath: "./testdata/http-server.log.sqlite",
		DoMigration:      true,
		Logger:           nullLogger,
	})
	require.Nil(t, err)

	t.Cleanup(func() {
		testDb.Close()
	})

	t.Run("can get specified row", func(t *testing.T) {
		query := SelectAllQuery(testDb, nullLogger)
		envelope := query.GetRow(4)
		assert.NotNil(t, envelope)
		assert.Equal(t, "Agent state changed from stopped to starting.", envelope.Message())
	})

	t.Run("search query returns expected rows", func(t *testing.T) {
		query := SearchQuery("shim", testDb, nullLogger)
		numRows := query.NumRows()
		assert.Equal(t, 385, numRows)

		rows, err := query.AllResults()
		assert.Nil(t, err)
		assert.Equal(t, 1, rows[0].RowId)
		assert.Equal(t, 385, rows[len(rows)-1].RowId)
	})
}
