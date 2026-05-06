package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Matthieusz/AVMS/internal/config"
	"github.com/Matthieusz/AVMS/internal/database"
	"github.com/Matthieusz/AVMS/internal/entry"
	"github.com/Matthieusz/AVMS/internal/server"
)

func initLogger(cfg config.Config) {
	format := cfg.Log.Format
	if format == "" {
		format = cfg.Server.GinMode
	}

	var handler slog.Handler
	if format == "json" || format == "release" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}

	slog.SetDefault(slog.New(handler))
}

func gracefulShutdown(srv *server.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	slog.Info("shutting down gracefully, press Ctrl+C again to force")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	if err := srv.Close(); err != nil {
		slog.Error("failed to close database", "error", err)
	}

	slog.Info("server exiting")
	done <- true
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	initLogger(cfg)

	db, err := database.Open(cfg.DB)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	migrator := database.NewMigrator()
	if err := migrator.Run(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	repo := database.New(db)
	entries := entry.NewService(repo)

	srv, err := server.New(cfg.Server, entries)
	if err != nil {
		slog.Error("failed to initialize server", "error", err)
		os.Exit(1)
	}

	done := make(chan bool, 1)
	go gracefulShutdown(srv, done)

	err = srv.Start()
	if err != nil && err != http.ErrServerClosed {
		slog.Error("http server error", "error", err)
		os.Exit(1)
	}

	<-done
	slog.Info("graceful shutdown complete")
}
