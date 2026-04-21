package main

import (
	"log"

	"github.com/sudo-odner/minor/backend/services/chat_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/chat_service/internal/lib/logger"
)

func main() {
	cfg := config.MustLoad()

	logger, err := logger.New(logger.Env(cfg.Env))
	if err != nil {
		log.Fatal("ERROR: failed initilizate logger: %s", err.Error())
	}
}
