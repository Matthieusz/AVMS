package server

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/Matthieusz/AVMS/internal/entry"
	"github.com/Matthieusz/AVMS/internal/pqc"
)

const maxCreateEntryBodySize = 1024 * 1024

type createEntryRequest struct {
	Value string `json:"value"`
}

func (s *Server) RegisterRoutes() http.Handler {
	if s.cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestIDMiddleware())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     s.cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		api.GET("/", s.HelloWorldHandler)
		api.GET("/health", s.healthHandler)
		api.GET("/health/detail", s.healthDetailHandler)
		api.GET("/pqc/kem-check", s.kemCheckHandler)
		api.GET("/entries", s.listEntriesHandler)
		api.POST("/entries", s.createEntryHandler)
		api.DELETE("/entries/:id", s.deleteEntryHandler)
	}

	// Serve frontend static files when a dist directory is configured.
	staticDir := s.cfg.StaticDir
	if stat, err := os.Stat(staticDir); err == nil && stat.IsDir() {
		r.Use(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				return
			}
			path := staticDir + c.Request.URL.Path
			info, err := os.Stat(path)
			if err == nil && !info.IsDir() {
				c.File(path)
				c.Abort()
			}
		})
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.File(staticDir + "/index.html")
		})
	}

	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "up"})
}

func (s *Server) healthDetailHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.entries.Health())
}

func (s *Server) createEntryHandler(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCreateEntryBodySize)

	var payload createEntryRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body is too large"})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	ent, err := s.entries.CreateEntry(c.Request.Context(), payload.Value)
	if err != nil {
		var ve entry.ValidationError
		if errors.As(err, &ve) {
			c.JSON(http.StatusBadRequest, gin.H{"error": ve.Message})
			return
		}

		logServerError(c, "create entry", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create entry"})
		return
	}

	c.JSON(http.StatusCreated, ent)
}

func (s *Server) listEntriesHandler(c *gin.Context) {
	entries, err := s.entries.ListEntries(c.Request.Context())
	if err != nil {
		logServerError(c, "list entries", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list entries"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

func (s *Server) deleteEntryHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry id"})
		return
	}

	deleted, err := s.entries.DeleteEntry(c.Request.Context(), id)
	if err != nil {
		logServerError(c, "delete entry", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete entry"})
		return
	}

	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (s *Server) kemCheckHandler(c *gin.Context) {
	result, err := pqc.RunKEMCheck("ML-KEM-512")
	if err != nil {
		logServerError(c, "run KEM check", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run KEM check"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func logServerError(c *gin.Context, operation string, err error) {
	logger, ok := c.Get("logger")
	if !ok {
		logger = slog.Default()
	}

	l, ok := logger.(*slog.Logger)
	if !ok {
		l = slog.Default()
	}

	l.Error(operation+" failed",
		"method", c.Request.Method,
		"path", c.FullPath(),
		"remote", c.ClientIP(),
		"error", err,
	)
}
