package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sudo-odner/minor/backend/service/dm_service/internal/config"
	"github.com/sudo-odner/minor/backend/service/dm_service/internal/repository/postgres"
)

type App struct {
	log *slog.Logger

	postgrespool *pgxpool.Pool
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	const op = "app.New"
	ctx := context.Background()
	a := &App{log: log}

	// Init repository
	pool, err := pgxpool.New(ctx, cfg.Postgres.ConnString())
	if err != nil {
		return nil, fmt.Errorf("%s: falied connect to postgres: %w", op, err)
	}
	a.postgrespool = pool
	_ = postgres.New(pool)

	return a, nil
}

func (a *App) Run() error {
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	if a.postgrespool != nil {
		a.postgrespool.Close()
	}
	return nil
}
