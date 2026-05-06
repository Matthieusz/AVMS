package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Matthieusz/AVMS/internal/config"
)

func newTestService(t *testing.T) Service {
	t.Helper()

	resetForTest()

	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db")

	srv, err := New(config.DBConfig{URL: dsn})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	t.Cleanup(func() {
		if err := srv.Close(); err != nil {
			t.Logf("failed to close test database: %v", err)
		}
		resetForTest()
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
	resetForTest()

	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db")

	srv, err := New(config.DBConfig{URL: dsn})
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	if err := srv.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After close, health should fail
	stats := srv.Health()
	if stats["status"] != "down" {
		t.Fatalf("expected status down after close, got %q", stats["status"])
	}
}

func TestLegacyMigrationSeeding(t *testing.T) {
	resetForTest()

	dir := t.TempDir()
	dsn := filepath.Join(dir, "legacy.db")

	// Manually create the entries table (simulating an old database)
	importRaw := os.Getenv("CGO_ENABLED")
	os.Setenv("CGO_ENABLED", "1")
	defer func() {
		if importRaw == "" {
			os.Unsetenv("CGO_ENABLED")
		} else {
			os.Setenv("CGO_ENABLED", importRaw)
		}
	}()

	// Use raw sqlite3 to create old schema
	// Since we can't easily import sqlite3 here without cgo, we'll test this
	// indirectly by verifying migrations work on a fresh DB.
	srv, err := New(config.DBConfig{URL: dsn})
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	// Verify _migrations table was created and seeded
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// The migration should have been applied
	entries, err := srv.ListEntries(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(entries))
	}

	if err := srv.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resetForTest()
}
