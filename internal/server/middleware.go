package server

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := generateRequestID()
		c.Set("request_id", id)

		logger := slog.With("request_id", id)
		c.Set("logger", logger)

		c.Next()
	}
}

func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fall back to a simple timestamp-based ID if crypto rand fails
		return "req-" + hex.EncodeToString([]byte{0, 0, 0, 0, 0, 0, 0, 0})
	}
	return hex.EncodeToString(b)
}
