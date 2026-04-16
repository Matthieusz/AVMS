package api

import (
	"fmt"

	"github.com/Kibaru/go-level-structure/internal/health"
	"github.com/Kibaru/go-level-structure/pkg/config"
	"github.com/gin-gonic/gin"
)

func NewServer() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	health.RegisterRoutes(api)

	return r
}

func Start() {
	r := NewServer()
	addr := fmt.Sprintf(":%s", config.Envs.Port)
	r.Run(addr)
}
