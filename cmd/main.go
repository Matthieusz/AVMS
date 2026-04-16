package main

import (
	"github.com/Kibaru/go-level-structure/cmd/api"
	"github.com/Kibaru/go-level-structure/pkg/config"
	"github.com/Kibaru/go-level-structure/pkg/logger"
)

func main() {
	logger.Init()
	logger.Info("Starting AVMS API server on port " + config.Envs.Port)
	api.Start()
}
