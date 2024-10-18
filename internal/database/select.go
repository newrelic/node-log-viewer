package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jsumners-nr/nr-node-logviewer/internal/common"
	v0 "github.com/jsumners-nr/nr-node-logviewer/internal/v0"
)

const selectSql = `
select rowid, original
from logs_fts
order by rowid asc
`

type SelectResult struct {
	Lines      []common.Envelope
	StartRowId int
	EndRowId   int
}

func (l *LogsDatabase) GetAllLogs() (*SelectResult, error) {
	return l.Select(0)
}

func (l *LogsDatabase) Select(limit int) (*SelectResult, error) {
	namedParams := make([]any, 0)
	statement := selectSql
	if limit > 0 {
		statement += "\nlimit @limit"
		namedParams = append(namedParams, sql.Named("limit", limit))
	}

	rows, err := l.Connection.Query(
		selectSql,
		namedParams...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	startRow := &DbRow{}
	endRow := &DbRow{}
	lines := make([]common.Envelope, 0)
	for rows.Next() {
		var row DbRow
		err = l.scanner.ScanRow(&row, rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

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
		lines = append(lines, envelope)
	}

	return &SelectResult{
		Lines:      lines,
		StartRowId: startRow.RowId,
		EndRowId:   endRow.RowId,
	}, nil
}
