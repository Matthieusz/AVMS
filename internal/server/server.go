package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"template/internal/database"
)

type Server struct {
	port int

	db database.Service
}

const defaultPort = 8080

func NewServer() (*http.Server, error) {
	port, err := portFromEnv()
	if err != nil {
		return nil, err
	}

	db, err := database.New()
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	newServer := &Server{
		port: port,

		db: db,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server, nil
}

func portFromEnv() (int, error) {
	return resolvePort(os.Getenv("PORT"))
}

func resolvePort(rawValue string) (int, error) {
	value := strings.TrimSpace(rawValue)
	if value == "" {
		return defaultPort, nil
	}

	port, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid PORT value %q: %w", value, err)
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("PORT out of range: %d", port)
	}

	return port, nil
}
