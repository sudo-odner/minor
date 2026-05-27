package main

import (
	"log"

	"github.com/sudo-odner/minor/backend/services/community_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/lib/logger"
)

func main() {
	// Init config and logger
	config := config.MustLoad()
	logger, err := logger.New(logger.Config{
		Env:         logger.Env(config.Env),
		ServiceName: "community-service",
	})
	if err != nil {
		log.Fatalf("FATAL: falied init logger: %s", err)
	}

	// Init application
	logger.Info("starting init application")
	application, err := 

	// TODO: Init application

	// TODO: Run and Stop app with graceful shutdown
}
