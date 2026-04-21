package main

import (
	"log"

	"github.com/sudo-odner/minor/backend/services/chat_service/internal/app"
	"github.com/sudo-odner/minor/backend/services/chat_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/chat_service/internal/lib/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.MustLoad()

	logger, err := logger.New(logger.Env(cfg.Env))
	if err != nil {
		log.Fatalf("ERROR: failed initilizate logger: %s", err.Error())
	}

	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Error("falied initilizate application", zap.Error(err))
		return
	}

	application.Run()
}
