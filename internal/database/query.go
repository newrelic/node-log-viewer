// As of 2026-01-08, the SQL statements in these methods suffer from full table
// scans. This is due to needing to map tview's `GetCel(rowNum, colNum)` method
// to a reproducible incrementing row identifier. The native sqlite rowid is
// not sufficient since we support filtering the full set of logs. Every set
// of rows returned from the database need incrementing ids starting from 1.
//
// Ultimately, we may need to avoid using tview's virtual table mechanism and
// craft our own table instances for each view.

package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/golang-lru/arc/v2"
	"github.com/newrelic/node-log-viewer/internal/common"
	"github.com/newrelic/node-log-viewer/internal/log"
	v0 "github.com/newrelic/node-log-viewer/internal/v0"
)

type Query struct {
	db           *LogsDatabase
	logger       *log.Logger
	rowCache     *arc.ARCCache[int, common.Envelope]
	materialized bool
	numRows      int
	text         string
}

// AllResults issues the base query statement and returns the set of
// [database.DbRow].
func (q *Query) AllResults() ([]DbRow, error) {
	if q.materialized == false {
		err := q.materialize()
		if err != nil {
			return nil, err
		}
	}

	rows, err := q.db.Connection.Query(`select rowid, * from mv`)
	if err != nil {
		return nil, fmt.Errorf("failed to query for all records: %w", err)
	}

	var dbRows []DbRow
	err = q.db.scanner.ScanAll(&dbRows, rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan query results: %w", err)
	}

	return dbRows, nil
}

func (q *Query) GetRow(number int) common.Envelope {
	if q.materialized == false {
		err := q.materialize()
		if err != nil {
			q.logger.Error("could not fetch requested row", "error", err)
			return nil
		}
	}

	if q.rowCache.Contains(number) {
		v, _ := q.rowCache.Get(number)
		return v
	}

	statement := fmt.Sprintf(
		`select * from mv where row_num = %d`,
		number,
	)

	rows, err := q.db.Connection.Query(statement)
	if err != nil {
		q.logger.Error("could not query for requested row", "error", err, "statement", statement)
		return nil
	}

	var dbRow DbRow
	err = q.db.scanner.ScanOne(&dbRow, rows)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		q.logger.Error("could not fetch requested row", "statement", statement)
		return nil
	case err != nil:
		q.logger.Error("failed to query for row", "error", err, "statement", statement)
		return nil
	}

	var envelope *v0.LineEnvelope
	err = json.Unmarshal([]byte(dbRow.Original), &envelope)
	if err != nil {
		q.logger.Error("failed to parse original log line", "error", err, "original", dbRow.Original)
		return nil
	}

	q.rowCache.Add(number, envelope)
	return envelope
}

func (q *Query) NumRows() int {
	if q.materialized == false {
		err := q.materialize()
		if err != nil {
			q.logger.Error("cannot determine number of rows", "error", err)
			return q.numRows
		}
	}

	if q.numRows > 0 {
		q.logger.Trace("returning cached number of rows", "numRows", q.numRows)
		return q.numRows
	}

	q.logger.Trace("querying for number of rows in view")
	statement := "select count(*) from mv"
	var numRows int
	err := q.db.Connection.QueryRow(statement).Scan(&numRows)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		q.logger.Error("could not fetch number of rows", "statement", statement)
		return 0
	case err != nil:
		q.logger.Error("failed to query for row count", "error", err, "statement", statement)
		return 0
	}

	q.logger.Trace("finished querying for number of rows in view", "numRows", numRows)
	q.numRows = numRows
	return numRows
}

func (q *Query) materialize() error {
	statement := fmt.Sprintf(
		`
			drop table if exists mv;
			drop index if exists mv_row_idx;
			create table mv as %s;
			create index mv_row_idx on mv (row_num);
		`,
		q.text,
	)
	_, err := q.db.Connection.Exec(statement)
	if err != nil {
		return fmt.Errorf("failed to materialize query: %w", err)
	}
	q.materialized = true
	return nil
}

func SelectAllQuery(db *LogsDatabase, logger *log.Logger) *Query {
	cache, _ := arc.NewARC[int, common.Envelope](1_024)
	return &Query{
		db:       db,
		logger:   logger,
		rowCache: cache,
		text:     `select row_number() over (order by rowid) as row_num, * from logs_fts`,
	}
}

func SearchQuery(searchTerm string, db *LogsDatabase, logger *log.Logger) *Query {
	cache, _ := arc.NewARC[int, common.Envelope](1_024)
	statement := fmt.Sprintf(
		`
			select
				row_number() over (order by rowid) as row_num, *
			from logs_fts
			where logs_fts match '%s'
		`,
		searchTerm,
	)
	return &Query{
		db:       db,
		logger:   logger,
		rowCache: cache,
		text:     statement,
	}
}
