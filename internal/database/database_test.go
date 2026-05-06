package database

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/Matthieusz/AVMS/internal/config"
)

func newTestService(t *testing.T) Service {
	t.Helper()

	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db")

	db, err := Open(config.DBConfig{URL: dsn})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	migrator := NewMigrator()
	if err := migrator.Run(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	srv := New(db)

	t.Cleanup(func() {
		if err := srv.Close(); err != nil {
			t.Logf("failed to close test database: %v", err)
		}
	})

	return srv
}

func TestHealth(t *testing.T) {
	srv := newTestService(t)

	stats := srv.Health()

	if stats["status"] != "up" {
		t.Fatalf("expected status up, got %q", stats["status"])
	}

	if stats["service"] != "api" {
		t.Fatalf("expected service api, got %q", stats["service"])
	}
}

func TestCreateEntry(t *testing.T) {
	srv := newTestService(t)
	ctx := context.Background()

	entry, err := srv.CreateEntry(ctx, "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	if entry.Value != "hello" {
		t.Fatalf("unexpected value: got %q, want %q", entry.Value, "hello")
	}

	if entry.CreatedAt == "" {
		t.Fatal("expected created_at to be set")
	}
}

func TestCreateEntryBlankValue(t *testing.T) {
	srv := newTestService(t)
	ctx := context.Background()

	// Database layer does not enforce blank value validation;
	// that is handled at the application layer.
	entry, err := srv.CreateEntry(ctx, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Value != "" {
		t.Fatalf("unexpected value: got %q, want empty string", entry.Value)
	}
}

func TestListEntries(t *testing.T) {
	srv := newTestService(t)
	ctx := context.Background()

	// Empty list
	entries, err := srv.ListEntries(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}

	// Create entries
	if _, err := srv.CreateEntry(ctx, "first"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := srv.CreateEntry(ctx, "second"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err = srv.ListEntries(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Verify descending order by ID
	if entries[0].Value != "second" {
		t.Fatalf("expected first entry to be 'second', got %q", entries[0].Value)
	}
	if entries[1].Value != "first" {
		t.Fatalf("expected second entry to be 'first', got %q", entries[1].Value)
	}
}

func TestDeleteEntry(t *testing.T) {
	srv := newTestService(t)
	ctx := context.Background()

	// Delete non-existent
	deleted, err := srv.DeleteEntry(ctx, 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted {
		t.Fatal("expected deleted to be false")
	}

	// Create and delete
	entry, err := srv.CreateEntry(ctx, "to-delete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	deleted, err = srv.DeleteEntry(ctx, entry.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatal("expected deleted to be true")
	}

	// Verify gone
	entries, err := srv.ListEntries(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries after delete, got %d", len(entries))
	}
}

func TestClose(t *testing.T) {
	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db")

	db, err := Open(config.DBConfig{URL: dsn})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	migrator := NewMigrator()
	if err := migrator.Run(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	srv := New(db)

	if err := srv.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After close, health should fail
	stats := srv.Health()
	if stats["status"] != "down" {
		t.Fatalf("expected status down after close, got %q", stats["status"])
	}
}

func TestMigratorSeedsLegacyMigration(t *testing.T) {
	dir := t.TempDir()
	dsn := filepath.Join(dir, "legacy.db")

	// Simulate an old database: create the entries table manually
	rawDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatalf("failed to open raw db: %v", err)
	}

	if _, err := rawDB.Exec(`
		CREATE TABLE entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			value TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		);
	`); err != nil {
		t.Fatalf("failed to create entries table: %v", err)
	}

	if err := rawDB.Close(); err != nil {
		t.Fatalf("failed to close raw db: %v", err)
	}

	// Now open via our package and run migrations
	db, err := Open(config.DBConfig{URL: dsn})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	migrator := NewMigrator()
	if err := migrator.Run(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Verify the legacy migration was seeded
	var exists int
	err = db.QueryRow("SELECT 1 FROM _migrations WHERE version = ?", "001_create_entries.sql").Scan(&exists)
	if err != nil {
		t.Fatalf("expected legacy migration to be seeded: %v", err)
	}
}
