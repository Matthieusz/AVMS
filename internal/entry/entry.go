package entry

import (
	"context"
	"errors"
)

// ValidationError is returned when input fails business-rule validation.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// Entry is the domain type for a stored record.
type Entry struct {
	ID        int64  `json:"id"`
	Value     string `json:"value"`
	CreatedAt string `json:"createdAt"`
}

// Service is the application-layer interface for entry operations.
type Service interface {
	// Health returns the underlying repository health status.
	Health() map[string]string

	// CreateEntry validates input and inserts a new record.
	CreateEntry(ctx context.Context, value string) (Entry, error)

	// ListEntries returns all stored records.
	ListEntries(ctx context.Context) ([]Entry, error)

	// DeleteEntry removes a record by ID.
	DeleteEntry(ctx context.Context, id int64) (bool, error)

	// Close terminates the underlying repository connection.
	Close() error
}

var (
	// ErrBlankValue is returned when the input value is empty or whitespace-only.
	ErrBlankValue = errors.New("value is required")
)
