package entry

import (
	"context"
	"strings"
	"time"

	"github.com/Matthieusz/AVMS/internal/database"
)

// repository is the narrow interface the service needs from storage.
type repository interface {
	Health() map[string]string
	CreateEntry(ctx context.Context, value string) (database.Entry, error)
	ListEntries(ctx context.Context) ([]database.Entry, error)
	DeleteEntry(ctx context.Context, id int64) (bool, error)
	Close() error
}

type service struct {
	repo repository
}

// NewService creates an entry Service backed by the given repository.
func NewService(repo repository) Service {
	return &service{repo: repo}
}

func (s *service) Health() map[string]string {
	return s.repo.Health()
}

func (s *service) CreateEntry(ctx context.Context, value string) (Entry, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Entry{}, ValidationError{Field: "value", Message: ErrBlankValue.Error()}
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	dbEntry, err := s.repo.CreateEntry(ctx, value)
	if err != nil {
		return Entry{}, err
	}

	return Entry{
		ID:        dbEntry.ID,
		Value:     dbEntry.Value,
		CreatedAt: dbEntry.CreatedAt,
	}, nil
}

func (s *service) ListEntries(ctx context.Context) ([]Entry, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	dbEntries, err := s.repo.ListEntries(ctx)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, len(dbEntries))
	for i, e := range dbEntries {
		entries[i] = Entry{
			ID:        e.ID,
			Value:     e.Value,
			CreatedAt: e.CreatedAt,
		}
	}

	return entries, nil
}

func (s *service) DeleteEntry(ctx context.Context, id int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return s.repo.DeleteEntry(ctx, id)
}

func (s *service) Close() error {
	return s.repo.Close()
}
