package main

import (
	"log"

	"github.com/sudo-odner/minor/backend/services/presence_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/lib/logger"
)

func main() {
	// Init config and logger
	cfg := config.MustLoad()
	logg, err := logger.New(logger.Config{
		Env:         logger.Env(cfg.Env),
		ServiceName: "community-service",
	})
	if err != nil {
		log.Fatalf("FATAL: failed init logger: %s", err)
	}
	_ = logg
}
