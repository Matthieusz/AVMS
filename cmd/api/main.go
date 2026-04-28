package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Matthieusz/AVMS/internal/server"
)

func initLogger() {
	format := os.Getenv("AVMS_LOG_FORMAT")
	if format == "" {
		format = os.Getenv("GIN_MODE")
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
	initLogger()

	srv, err := server.NewServer()
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
