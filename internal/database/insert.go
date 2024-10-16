package database

import (
	"database/sql"
	v0 "github.com/jsumners-nr/nr-node-logviewer/internal/v0"
)

const insertSql = `
	insert into logs (version, time, component, message, original)
	values (@version, @time, @component, @message, @original)
`

func (l *LogsDatabase) Insert(log *v0.LineEnvelope, source string) error {
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
