package app

import (
	"context"
	"log/slog"

	"github.com/sudo-odner/minor/backend/service/dm_service/internal/config"
)

type App struct{}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	return nil, nil
}

func (a *App) Run() error {
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	return nil
}
