package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/newrelic/node-log-viewer/internal/common"
	v0 "github.com/newrelic/node-log-viewer/internal/v0"
)

const selectSql = `
select rowid, *
from logs_fts
`

type SelectedRow struct {
	common.Envelope
	RowId    int
	Original string
}

type SelectResult struct {
	Rows       []SelectedRow
	StartRowId int
	EndRowId   int
}

// ToLines converts the selection result into a set of parsed [common.Envelope]
// objects.
func (s *SelectResult) ToLines() []common.Envelope {
	lines := make([]common.Envelope, 0, len(s.Rows))
	for _, line := range s.Rows {
		lines = append(lines, line.Envelope)
	}
	return lines
}

func (l *LogsDatabase) GetAllLogs() (*SelectResult, error) {
	return l.Select(0, "")
}

func (l *LogsDatabase) Search(term string) (*SelectResult, error) {
	if term == "" {
		return l.GetAllLogs()
	}
	return l.Select(0, fmt.Sprintf("where logs_fts match '%s'", term))
}

func (l *LogsDatabase) Select(limit int, clause string) (*SelectResult, error) {
	l.logger.Trace("performing database select", "limit", limit, "clause", clause)
	namedParams := make([]any, 0)
	statement := selectSql

	if clause != "" {
		statement += "\n" + clause + "\n"
	}

	if limit > 0 {
		statement += "\nlimit @limit"
		namedParams = append(namedParams, sql.Named("limit", limit))
	}

	rows, err := l.Connection.Query(
		statement,
		namedParams...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var startRow *DbRow
	var endRow *DbRow
	lines := make([]SelectedRow, 0)
	for rows.Next() {
		var row DbRow
		err = l.scanner.ScanRow(&row, rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// On the first pass, we set the first row as the starting row.
		// On all subsequent passes, we assign/re-assign the current row as the
		// ending row.
		if startRow == nil {
			startRow = &row
		} else {
			endRow = &row
		}

		var envelope *v0.LineEnvelope
		err = json.Unmarshal([]byte(row.Original), &envelope)
		if err != nil {
			return nil, fmt.Errorf("failed to parse original log line: %w", err)
		}
		lines = append(lines, SelectedRow{
			Envelope: envelope,
			RowId:    row.RowId,
			Original: row.Original,
		})
	}

	// If we only had a single row result, then we never had a second pass that
	// would have set an ending row.
	if endRow == nil {
		endRow = startRow
	}

	return &SelectResult{
		Rows:       lines,
		StartRowId: startRow.RowId,
		EndRowId:   endRow.RowId,
	}, nil
}
