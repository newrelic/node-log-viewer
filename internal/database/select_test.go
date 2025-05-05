package database

import (
	"github.com/gookit/goutil/testutil/assert"
	"github.com/newrelic/node-log-viewer/internal/common"
	"github.com/newrelic/node-log-viewer/internal/log"
	v0 "github.com/newrelic/node-log-viewer/internal/v0"
	"github.com/stretchr/testify/require"
	"testing"
)

var nullLogger = log.NewDiscardLogger()

func Test_ToLines(t *testing.T) {
	selectResult := SelectResult{
		Rows: []SelectedRow{
			{RowId: 1, Original: "foo", Envelope: &v0.LineEnvelope{Version: 1, Name: "foo"}},
			{RowId: 2, Original: "bar", Envelope: &v0.LineEnvelope{Version: 1, Name: "bar"}},
		},
		StartRowId: 1,
		EndRowId:   2,
	}
	found := selectResult.ToLines()
	expected := []common.Envelope{
		&v0.LineEnvelope{Version: 1, Name: "foo"},
		&v0.LineEnvelope{Version: 1, Name: "bar"},
	}
	assert.Equal(t, expected, found)
}

func Test_Select(t *testing.T) {
	testDb, err := New(DbParams{
		DatabaseFilePath: "./testdata/http-server.log.sqlite",
		DoMigration:      true,
		Logger:           nullLogger,
	})
	require.Nil(t, err)

	t.Cleanup(func() {
		testDb.Close()
	})

	t.Run("works with empty db", func(t *testing.T) {
		db, err := New(DbParams{
			DatabaseFilePath: "file::memory:",
			DoMigration:      true,
			Logger:           nullLogger,
		})
		require.Nil(t, err)

		results, err := db.Select(0, "")
		assert.Nil(t, err)
		assert.Equal(t, 0, len(results.Rows))
	})

	t.Run("finds limited set of rows for a search term", func(t *testing.T) {
		results, err := testDb.Select(1, "where logs_fts match 'shim'")
		assert.Nil(t, err)
		assert.Equal(t, 1, len(results.Rows))
		assert.Equal(t, 107, results.StartRowId)
		assert.Equal(t, 107, results.EndRowId)

		row := results.Rows[0]
		assert.Equal(t, 107, row.RowId)
		assert.Contains(t, row.Original, `"component":"Shim"`)
		envelope := row.Envelope.(*v0.LineEnvelope)
		assert.Equal(t, "newrelic", envelope.Name)
		assert.Equal(t, "Shim", envelope.SourceComponent)
	})

	t.Run("tracks end row correctly", func(t *testing.T) {
		results, err := testDb.Select(4, "where logs_fts match 'shim'")
		assert.Nil(t, err)
		assert.Equal(t, 4, len(results.Rows))
		assert.Equal(t, 107, results.StartRowId)
		assert.Equal(t, 110, results.EndRowId)
	})

	t.Run("finds all matched rows", func(t *testing.T) {
		results, err := testDb.Select(0, "where logs_fts match 'shim'")
		assert.Nil(t, err)
		assert.Equal(t, 385, len(results.Rows))
		assert.Equal(t, 107, results.StartRowId)
		assert.Equal(t, 8077, results.EndRowId)
	})

	t.Run("returns all rows for no search term", func(t *testing.T) {
		results, err := testDb.Select(0, "")
		assert.Nil(t, err)
		assert.Equal(t, 8092, len(results.Rows))
	})
}

func Test_Search(t *testing.T) {
	testDb, err := New(DbParams{
		DatabaseFilePath: "./testdata/http-server.log.sqlite",
		DoMigration:      true,
		Logger:           nullLogger,
	})
	require.Nil(t, err)

	t.Cleanup(func() {
		testDb.Close()
	})

	t.Run("gets all logs for empty term", func(t *testing.T) {
		results, err := testDb.Search("")
		assert.Nil(t, err)
		assert.Equal(t, 8092, len(results.Rows))
	})

	t.Run("gets correct logs for search term", func(t *testing.T) {
		results, err := testDb.Search("shim")
		assert.Nil(t, err)
		assert.Equal(t, 385, len(results.Rows))
		assert.Equal(t, 107, results.StartRowId)
		assert.Equal(t, 8077, results.EndRowId)
	})
}
