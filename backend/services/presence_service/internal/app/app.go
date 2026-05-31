package app

import (
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/config"
	"go.uber.org/zap"
)

type App struct {
	log *zap.Logger
}

func New(cfg *config.Config, log *zap.Logger) *App {
	return &App{
		log: log,
	}
}
