package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Matthieusz/AVMS/internal/config"
	"github.com/Matthieusz/AVMS/internal/entry"
)

// Server is the HTTP server module.
type Server struct {
	entries    entry.Service
	cfg        config.ServerConfig
	httpServer *http.Server
}

// New creates a Server wired with the given configuration and entry service.
func New(cfg config.ServerConfig, entries entry.Service) (*Server, error) {
	s := &Server{
		entries: entries,
		cfg:     cfg,
	}

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
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
	return s.entries.Close()
}
