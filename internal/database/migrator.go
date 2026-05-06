package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Migrator applies schema migrations to a SQLite database.
type Migrator struct {
	fs embed.FS
}

// NewMigrator creates a new Migrator using the embedded migration files.
func NewMigrator() *Migrator {
	return &Migrator{fs: migrationFS}
}

// Run creates the migrations tracking table, seeds any legacy migrations, and
// applies all pending migration files in lexicographic order.
func (m *Migrator) Run(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS _migrations (
			version TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		);
	`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Backwards compatibility: if entries table exists but no migration recorded,
	// record the first migration as already applied.
	if err := m.seedLegacyMigration(db); err != nil {
		return fmt.Errorf("seed legacy migration: %w", err)
	}

	entries, err := fs.ReadDir(m.fs, "migrations")
	if err != nil {
		return fmt.Errorf("read migration directory: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}

	sort.Strings(files)

	for _, version := range files {
		var exists int
		err := db.QueryRow("SELECT 1 FROM _migrations WHERE version = ?", version).Scan(&exists)
		if err == nil {
			// Already applied
			continue
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("check migration %q: %w", version, err)
		}

		sqlBytes, err := fs.ReadFile(m.fs, "migrations/"+version)
		if err != nil {
			return fmt.Errorf("read migration %q: %w", version, err)
		}

		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("apply migration %q: %w", version, err)
		}

		if _, err := db.Exec("INSERT INTO _migrations(version) VALUES (?)", version); err != nil {
			return fmt.Errorf("record migration %q: %w", version, err)
		}

		slog.Info("applied migration", "version", version)
	}

	return nil
}

func (m *Migrator) seedLegacyMigration(db *sql.DB) error {
	var entriesExists int
	err := db.QueryRow("SELECT 1 FROM sqlite_master WHERE type='table' AND name='entries'").Scan(&entriesExists)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("check entries table existence: %w", err)
	}

	if err == sql.ErrNoRows {
		// entries table does not exist yet; normal migration flow will create it
		return nil
	}

	var migrationExists int
	err = db.QueryRow("SELECT 1 FROM _migrations WHERE version = ?", "001_create_entries.sql").Scan(&migrationExists)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("check migration record: %w", err)
	}

	if err == sql.ErrNoRows {
		// entries table exists but migration not recorded; seed it
		if _, err := db.Exec("INSERT INTO _migrations(version) VALUES (?)", "001_create_entries.sql"); err != nil {
			return fmt.Errorf("seed legacy migration: %w", err)
		}
		slog.Info("seeded legacy migration", "version", "001_create_entries.sql")
	}

	return nil
}
