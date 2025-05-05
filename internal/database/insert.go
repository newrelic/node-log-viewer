package database

import (
	"database/sql"
	v0 "github.com/newrelic/node-log-viewer/internal/v0"
)

const insertSql = `
	insert into logs (version, time, component, message, original)
	values (@version, @time, @component, @message, @original)
`

type InsertTuple struct {
	ParsedLog *v0.LineEnvelope
	Source    string
}

func (l *LogsDatabase) Insert(tuple InsertTuple) error {
	log := tuple.ParsedLog
	source := tuple.Source
	_, err := l.Connection.Exec(
		insertSql,
		sql.Named("version", log.Version),
		sql.Named("time", log.Time),
		sql.Named("component", log.SourceComponent),
		sql.Named("message", log.LogMessage),
		sql.Named("original", source),
	)
	return err
}

func (l *LogsDatabase) BatchInsert(tuples []InsertTuple) error {
	l.logger.Debug("inserting batch of logs", "batch_size", len(tuples))
	for _, tuple := range tuples {
		err := l.Insert(tuple)
		if err != nil {
			l.logger.Error("failed to insert line into database", "error", err, "line", tuple.Source)
			return err
		}
	}
	return nil
}
