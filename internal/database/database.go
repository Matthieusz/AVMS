package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/Matthieusz/AVMS/internal/config"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// CreateEntry inserts a new record in the database.
	CreateEntry(ctx context.Context, value string) (Entry, error)

	// ListEntries returns all stored records.
	ListEntries(ctx context.Context) ([]Entry, error)

	// DeleteEntry removes a record by ID.
	DeleteEntry(ctx context.Context, id int64) (bool, error)

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error
}

// Entry is a stored record.
type Entry struct {
	ID        int64  `json:"id"`
	Value     string `json:"value"`
	CreatedAt string `json:"createdAt"`
}

type service struct {
	db *sql.DB
}

// Open opens the SQLite database described by cfg, applies connection pragmas,
// and returns the raw *sql.DB. The caller owns the lifecycle: run migrations,
// create a Service, and close the connection when done.
func Open(cfg config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database %q: %w", cfg.URL, err)
	}

	// SQLite-specific connection pool tuning
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(0)

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			if closeErr := db.Close(); closeErr != nil {
				slog.Error("failed to close database after pragma error", "error", closeErr)
			}
			return nil, fmt.Errorf("execute %q: %w", pragma, err)
		}
	}

	return db, nil
}

// New creates a Service backed by the given database connection.
func New(db *sql.DB) Service {
	return &service{db: db}
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		stats["service"] = "api"
		stats["timestamp"] = time.Now().UTC().Format(time.RFC3339)
		slog.Error("db down", "error", err)
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"
	stats["service"] = "api"
	stats["timestamp"] = time.Now().UTC().Format(time.RFC3339)

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

func (s *service) CreateEntry(ctx context.Context, value string) (Entry, error) {
	result, err := s.db.ExecContext(ctx, "INSERT INTO entries(value) VALUES (?)", value)
	if err != nil {
		return Entry{}, fmt.Errorf("failed to insert entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Entry{}, fmt.Errorf("failed to get inserted entry id: %w", err)
	}

	var entry Entry
	err = s.db.QueryRowContext(ctx, "SELECT id, value, created_at FROM entries WHERE id = ?", id).
		Scan(&entry.ID, &entry.Value, &entry.CreatedAt)
	if err != nil {
		return Entry{}, fmt.Errorf("failed to fetch inserted entry: %w", err)
	}

	return entry, nil
}

func (s *service) ListEntries(ctx context.Context) ([]Entry, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, value, created_at FROM entries ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to list entries: %w", err)
	}
	defer rows.Close()

	entries := make([]Entry, 0)
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.Value, &entry.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while iterating entries: %w", err)
	}

	return entries, nil
}

func (s *service) DeleteEntry(ctx context.Context, id int64) (bool, error) {
	result, err := s.db.ExecContext(ctx, "DELETE FROM entries WHERE id = ?", id)
	if err != nil {
		return false, fmt.Errorf("failed to delete entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to determine deleted rows: %w", err)
	}

	return rowsAffected > 0, nil
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	slog.Info("disconnected from database")
	return s.db.Close()
}
