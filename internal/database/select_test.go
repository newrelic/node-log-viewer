package database

import (
	"github.com/gookit/goutil/testutil/assert"
	"github.com/newrelic/node-log-viewer/internal/common"
	v0 "github.com/newrelic/node-log-viewer/internal/v0"
	"testing"
)

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
