package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/v2/dbscan"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/golang-migrate/migrate/v4"
	migrateSqlite "github.com/golang-migrate/migrate/v4/database/sqlite"
	migrateFS "github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jsumners-nr/nr-node-logviewer/internal/database/migrations"
	"github.com/jsumners-nr/nr-node-logviewer/internal/log"
	"github.com/jsumners/go-rfc3339"

	// We have to load the sqlite driver without using it because Go's stdlib
	// database system relies on import side effects for loading database drivers.
	// Yes, this is dumb.
	_ "modernc.org/sqlite"
)

type LogsDatabase struct {
	Connection   *sql.DB
	DatabaseFile string
	logger       *log.Logger
	scanner      *sqlscan.API
}

type DbParams struct {
	DatabaseFilePath string
	DoMigration      bool
	Logger           *log.Logger
}

type DbRow struct {
	RowId     int
	Version   int
	Time      rfc3339.DateTime
	Component string
	Message   string
	Original  string
}

func New(params DbParams) (*LogsDatabase, error) {
	// TODO: utilize options instead of a params struct
	// TODO: default to a discard logger
	result := &LogsDatabase{
		DatabaseFile: params.DatabaseFilePath,
		logger:       params.Logger,
	}

	db, err := sql.Open("sqlite", params.DatabaseFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database file: %w", err)
	}
	result.Connection = db

	dbScanApi, err := sqlscan.NewDBScanAPI(
		dbscan.WithAllowUnknownColumns(true),
	)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create scanner config: %w", err)
	}
	scanner, err := sqlscan.NewAPI(dbScanApi)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create scanner: %w", err)
	}
	result.scanner = scanner

	if params.DoMigration == true {
		err = result.MigrateUp()
		if err != nil {
			db.Close()
			return nil, err
		}
	}

	return result, nil
}

func (l *LogsDatabase) Close() {
	err := l.Connection.Close()
	if err != nil {
		l.logger.Error("error closing database", "error", err)
	}
}

func (l *LogsDatabase) MigrateUp() error {
	return migrateUp(l.Connection)
}

func migrateUp(db *sql.DB) error {
	// Set up the driver for the migration library:
	driver, err := migrateSqlite.WithInstance(db, &migrateSqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Define the file system for the migrator:
	fsDriver, err := migrateFS.New(migrations.FS, "sql")
	if err != nil {
		return fmt.Errorf("failed to setup migrations fs: %w", err)
	}

	// Run the migrations:
	migrator, err := migrate.NewWithInstance("iofs", fsDriver, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to create database migrator: %w", err)
	}
	err = migrator.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("failed to run database migration: %w", err)
	}

	return nil
}
