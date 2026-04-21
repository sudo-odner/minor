package app

import (
	"context"

	"github.com/sudo-odner/minor/backend/services/chat_service/internal/config"
	"go.uber.org/zap"
)

type App struct {
	log *zap.Logger
}

func New(cfg *config.Config, log *zap.Logger) (*App, error) {
	// TODO: repo cassandra
	// TODO: brocker nuts
	// TODO: services
	// TODO: handler

	return &App{
		log: log,
	}, nil
}

func (a *App) Run() error {
	const op = "app.Run"
	log := a.log.With(zap.String("op", op))

	log.Info("application started")
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	const op = "app.Stop"
	log := a.log.With(zap.String("op", op))

	log.Info("application stopped successfully")
	return nil
}
