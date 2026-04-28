package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/Matthieusz/AVMS/internal/database"
)

type Server struct {
	db         database.Service
	httpServer *http.Server
}

const defaultPort = 8080

func NewServer() (*Server, error) {
	port, err := portFromEnv()
	if err != nil {
		return nil, err
	}

	db, err := database.New()
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	s := &Server{db: db}

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return s, nil
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Close() error {
	return s.db.Close()
}

func portFromEnv() (int, error) {
	value := os.Getenv("AVMS_PORT")
	if value == "" {
		value = os.Getenv("PORT")
	}
	return resolvePort(value)
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
