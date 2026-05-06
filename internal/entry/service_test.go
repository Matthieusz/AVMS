package entry

import (
	"context"
	"errors"
	"testing"

	"github.com/Matthieusz/AVMS/internal/database"
)

type stubRepo struct {
	createEntryCalls int
	deleteEntryCalls int

	createEntryFunc func(ctx context.Context, value string) (database.Entry, error)
	listEntriesFunc func(ctx context.Context) ([]database.Entry, error)
	deleteEntryFunc func(ctx context.Context, id int64) (bool, error)
}

func (s *stubRepo) Health() map[string]string {
	return map[string]string{"status": "up"}
}

func (s *stubRepo) CreateEntry(ctx context.Context, value string) (database.Entry, error) {
	s.createEntryCalls++
	if s.createEntryFunc != nil {
		return s.createEntryFunc(ctx, value)
	}
	return database.Entry{ID: 1, Value: value, CreatedAt: "2026-01-01T00:00:00Z"}, nil
}

func (s *stubRepo) ListEntries(ctx context.Context) ([]database.Entry, error) {
	if s.listEntriesFunc != nil {
		return s.listEntriesFunc(ctx)
	}
	return []database.Entry{}, nil
}

func (s *stubRepo) DeleteEntry(ctx context.Context, id int64) (bool, error) {
	s.deleteEntryCalls++
	if s.deleteEntryFunc != nil {
		return s.deleteEntryFunc(ctx, id)
	}
	return true, nil
}

func (s *stubRepo) Close() error {
	return nil
}

func TestCreateEntryRejectsBlankValue(t *testing.T) {
	srv := NewService(&stubRepo{})
	ctx := context.Background()

	_, err := srv.CreateEntry(ctx, "   ")
	if err == nil {
		t.Fatal("expected error for blank value")
	}

	var ve ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if ve.Field != "value" {
		t.Fatalf("unexpected field: got %q, want %q", ve.Field, "value")
	}
}

func TestCreateEntryTrimsWhitespace(t *testing.T) {
	repo := &stubRepo{}
	srv := NewService(repo)
	ctx := context.Background()

	entry, err := srv.CreateEntry(ctx, "  hello  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.Value != "hello" {
		t.Fatalf("unexpected value: got %q, want %q", entry.Value, "hello")
	}

	if repo.createEntryCalls != 1 {
		t.Fatalf("expected 1 create call, got %d", repo.createEntryCalls)
	}
}

func TestCreateEntryPropagatesRepositoryError(t *testing.T) {
	repo := &stubRepo{
		createEntryFunc: func(_ context.Context, _ string) (database.Entry, error) {
			return database.Entry{}, errors.New("db down")
		},
	}
	srv := NewService(repo)
	ctx := context.Background()

	_, err := srv.CreateEntry(ctx, "hello")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListEntries(t *testing.T) {
	repo := &stubRepo{
		listEntriesFunc: func(_ context.Context) ([]database.Entry, error) {
			return []database.Entry{
				{ID: 1, Value: "first", CreatedAt: "2026-01-01T00:00:00Z"},
				{ID: 2, Value: "second", CreatedAt: "2026-01-02T00:00:00Z"},
			}, nil
		},
	}
	srv := NewService(repo)
	ctx := context.Background()

	entries, err := srv.ListEntries(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Value != "first" {
		t.Fatalf("unexpected first value: got %q", entries[0].Value)
	}
}

func TestDeleteEntry(t *testing.T) {
	repo := &stubRepo{}
	srv := NewService(repo)
	ctx := context.Background()

	deleted, err := srv.DeleteEntry(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatal("expected deleted to be true")
	}
	if repo.deleteEntryCalls != 1 {
		t.Fatalf("expected 1 delete call, got %d", repo.deleteEntryCalls)
	}
}

func TestDeleteEntryReturnsFalseWhenNotFound(t *testing.T) {
	repo := &stubRepo{
		deleteEntryFunc: func(_ context.Context, _ int64) (bool, error) {
			return false, nil
		},
	}
	srv := NewService(repo)
	ctx := context.Background()

	deleted, err := srv.DeleteEntry(ctx, 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted {
		t.Fatal("expected deleted to be false")
	}
}

func TestHealth(t *testing.T) {
	srv := NewService(&stubRepo{})
	stats := srv.Health()
	if stats["status"] != "up" {
		t.Fatalf("expected status up, got %q", stats["status"])
	}
}

func TestClose(t *testing.T) {
	srv := NewService(&stubRepo{})
	if err := srv.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
