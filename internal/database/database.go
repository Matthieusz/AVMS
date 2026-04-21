package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// CreateItem inserts a new record in the database.
	CreateItem(ctx context.Context, value string) (Item, error)

	// ListItems returns all stored records.
	ListItems(ctx context.Context) ([]Item, error)

	// DeleteItem removes a record by ID.
	DeleteItem(ctx context.Context, id int64) (bool, error)

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error
}

type Item struct {
	ID        int64  `json:"id"`
	Value     string `json:"value"`
	CreatedAt string `json:"createdAt"`
}

type service struct {
	db  *sql.DB
	dsn string
}

var (
	dbInstance *service
	dbMu       sync.Mutex
)

const defaultDBURL = "./test.db"

func New() (Service, error) {
	dbMu.Lock()
	defer dbMu.Unlock()

	// Reuse Connection
	if dbInstance != nil {
		return dbInstance, nil
	}

	dburl := databaseURL()
	db, err := sql.Open("sqlite3", dburl)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database %q: %w", dburl, err)
	}

	instance := &service{
		db:  db,
		dsn: dburl,
	}

	if err := instance.initSchema(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("failed to close database after schema init error: %v", closeErr)
		}
		return nil, err
	}

	dbInstance = instance

	return dbInstance, nil
}

func databaseURL() string {
	value := strings.TrimSpace(os.Getenv("BLUEPRINT_DB_URL"))
	if value == "" {
		return defaultDBURL
	}

	return value
}

func (s *service) initSchema() error {
	const schema = `
		CREATE TABLE IF NOT EXISTS entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			value TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		);
	`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
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
		log.Printf("db down: %v", err)
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

func (s *service) CreateItem(ctx context.Context, value string) (Item, error) {
	result, err := s.db.ExecContext(ctx, "INSERT INTO entries(value) VALUES (?)", value)
	if err != nil {
		return Item{}, fmt.Errorf("failed to insert item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Item{}, fmt.Errorf("failed to get inserted item id: %w", err)
	}

	var item Item
	err = s.db.QueryRowContext(ctx, "SELECT id, value, created_at FROM entries WHERE id = ?", id).
		Scan(&item.ID, &item.Value, &item.CreatedAt)
	if err != nil {
		return Item{}, fmt.Errorf("failed to fetch inserted item: %w", err)
	}

	return item, nil
}

func (s *service) ListItems(ctx context.Context) ([]Item, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, value, created_at FROM entries ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}
	defer rows.Close()

	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Value, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while iterating items: %w", err)
	}

	return items, nil
}

func (s *service) DeleteItem(ctx context.Context, id int64) (bool, error) {
	result, err := s.db.ExecContext(ctx, "DELETE FROM entries WHERE id = ?", id)
	if err != nil {
		return false, fmt.Errorf("failed to delete item: %w", err)
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
	log.Printf("Disconnected from database: %s", s.dsn)
	return s.db.Close()
}
