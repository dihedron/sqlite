package sqlite

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/mattn/go-sqlite3" // SQLite3 drivers

	"github.com/dihedron/sqlite/log"
	"go.uber.org/zap"
)

// InitDB opens and initalises an SQLite3 DB with all
// correct settings.
func InitDB(dsn string, migrations fs.FS) (db *sql.DB, err error) {
	// ensure a DSN is set before attempting to open the database
	if dsn == "" {
		log.L.Error("the database DSN must be specified")
		err = fmt.Errorf("dsn required")
		return
	}
	log.L.Debug("opening SQLite3 database", zap.String("DSN", dsn))
	// make the parent directory unless using an in-memory db
	if dsn != ":memory:" {
		if err = os.MkdirAll(filepath.Dir(dsn), 0700); err != nil {
			log.L.Error("error creating directory for on-disk DB file", zap.String("path", dsn), zap.Error(err))
			return
		}
	}
	// open the database
	if db, err = sql.Open("sqlite3", dsn); err != nil {
		log.L.Error("error connecting to the database", zap.Error(err))
		return
	}
	// enable WAL; SQLite performs better with the WAL because it allows
	// multiple readers to operate while data is being written
	if _, err = db.Exec(`PRAGMA journal_mode = wal;`); err != nil {
		log.L.Error("error enabling WAL", zap.Error(err))
		err = fmt.Errorf("enable wal: %w", err)
		return
	}
	// enable foreign key checks: for historical reasons, SQLite does not check
	// foreign key constraints by default... which is kinda insane; there's some
	// overhead on inserts to verify foreign key integrity but it's definitely
	// worth it.
	if _, err = db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		log.L.Error("error enabling foreign keys checks", zap.Error(err))
		err = fmt.Errorf("foreign keys pragma: %w", err)
		return
	}

	// apply migrations (if any)
	if err = migrate(db, migrations); err != nil {
		log.L.Error("error applying migrations", zap.Error(err))
		err = fmt.Errorf("migrate: %w", err)
		return
	}
	return
}

// migrate sets up migration tracking and executes pending migration files.
//
// Migration files can be on disk or embedded in the executable. This function
// gets a list and sorts it so that they are executed in lexigraphical order.
//
// Once a migration is run, its name is stored in the 'migrations' table so it
// is not re-executed. Migrations run in a transaction to prevent partial
// migrations.
func migrate(db *sql.DB, migrations fs.FS) error {
	log.L.Debug("applying migrations...")

	// ensure the 'migrations' table exists so we don't duplicate migrations.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS migrations (name TEXT PRIMARY KEY);`); err != nil {
		log.L.Error("error creating migrations table", zap.Error(err))
		return fmt.Errorf("cannot create migrations table: %w", err)
	}

	// read migration files from our embedded file system;
	// this uses Go 1.16's 'embed' package.
	names, err := fs.Glob(migrations, "*.sql")
	if err != nil {
		log.L.Error("error getting list of migrations", zap.Error(err))
		return err
	}
	sort.Strings(names)

	// loop over all migration files and execute them in order
	for _, name := range names {
		if err := migrateFile(db, migrations, name); err != nil {
			log.L.Error("error applying migration file", zap.String("name", name), zap.Error(err))
			return fmt.Errorf("migration error: name=%q err=%w", name, err)
		}
	}
	log.L.Debug("all migrations applied")
	return nil
}

// migrate runs a single migration file within a transaction; on success, the
// migration file name is saved to the "migrations" table to prevent re-running.
func migrateFile(db *sql.DB, migrations fs.FS, name string) error {
	log.L.Debug("applying migration file", zap.String("name", name))
	tx, err := db.Begin()
	if err != nil {
		log.L.Error("error opening transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	// ensure migration has not already been run
	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM migrations WHERE name = ?`, name).Scan(&n); err != nil {
		log.L.Error("error reading migrations table", zap.String("name", name), zap.Error(err))
		return err
	} else if n != 0 {
		log.L.Debug("mmigration already applied, skipping", zap.String("name", name))
		return nil // already run migration, skip
	}

	// read and execute migration file
	if buffer, err := fs.ReadFile(migrations, name); err != nil {
		log.L.Error("error reading migration file", zap.String("name", name), zap.Error(err))
		return err
	} else if _, err := tx.Exec(string(buffer)); err != nil {
		log.L.Error("error executing migration command", zap.String("command", string(buffer)), zap.Error(err))
		return err
	}

	// insert record into migrations to prevent re-running migration
	if _, err := tx.Exec(`INSERT INTO migrations (name) VALUES (?)`, name); err != nil {
		log.L.Error("error inserting migration into migrations table", zap.String("name", name))
		return err
	}
	log.L.Debug("migration applied", zap.String("name", name))
	return tx.Commit()
}
