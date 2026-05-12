package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/Matthieusz/AVMS/internal/database"
	"github.com/Matthieusz/AVMS/internal/pqc"
)

const (
	defaultAllowedOrigin  = "http://localhost:5173"
	maxCreateItemBodySize = 1024 * 1024
)

type createItemRequest struct {
	Value string `json:"value"`
}

func (s *Server) RegisterRoutes() http.Handler {
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestIDMiddleware())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOriginsFromEnv(),
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
		api.GET("/rsus/:rsuID/beacon", s.rsuBeaconHandler)
		api.POST("/vehicles/register", s.registerVehicleHandler)
		api.POST("/vehicles/join", s.joinVehicleHandler)
		api.POST("/certificates/issue", s.issueCredentialHandler)
		api.POST("/certificates/revoke", s.revokeCredentialHandler)
		api.POST("/keys/rotate", s.rotateKeysHandler)
		api.GET("/credentials/:credentialID/status", s.credentialStatusHandler)
		api.GET("/policies/current", s.currentPolicyHandler)
		api.POST("/incidents/report", s.reportIncidentHandler)
		api.GET("/items", s.listItemsHandler)
		api.POST("/items", s.createItemHandler)
		api.DELETE("/items/:id", s.deleteItemHandler)
	}

	// Serve frontend static files when a dist directory is configured.
	staticDir := os.Getenv("AVMS_STATIC_DIR")
	if staticDir == "" {
		staticDir = "./frontend/dist"
	}
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
	if s.db == nil {
		c.JSON(http.StatusOK, gin.H{"status": "up"})
		return
	}

	c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) healthDetailHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) createItemHandler(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCreateItemBodySize)

	var payload createItemRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body is too large"})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	value := strings.TrimSpace(payload.Value)
	if value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	item, err := s.db.CreateItem(ctx, value)
	if err != nil {
		logServerError(c, "create item", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create item"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (s *Server) listItemsHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	items, err := s.db.ListItems(ctx)
	if err != nil {
		logServerError(c, "list items", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) deleteItemHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	deleted, err := s.db.DeleteItem(ctx, id)
	if err != nil {
		logServerError(c, "delete item", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete item"})
		return
	}

	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (s *Server) kemCheckHandler(c *gin.Context) {
	result, err := pqc.RunKEMCheck(database.RecommendedKEMAlgorithm)
	if err != nil {
		logServerError(c, "run KEM check", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run KEM check"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func allowedOriginsFromEnv() []string {
	value := strings.TrimSpace(os.Getenv("AVMS_CORS_ORIGINS"))
	if value == "" {
		value = strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	}
	return parseAllowedOrigins(value)
}

func parseAllowedOrigins(raw string) []string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return []string{defaultAllowedOrigin}
	}

	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))

	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" || origin == "*" {
			continue
		}

		if _, exists := seen[origin]; exists {
			continue
		}

		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}

	if len(origins) == 0 {
		return []string{defaultAllowedOrigin}
	}

	return origins
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
