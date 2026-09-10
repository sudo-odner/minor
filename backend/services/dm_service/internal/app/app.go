package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sudo-odner/minor/backend/service/dm_service/internal/config"
	repo "github.com/sudo-odner/minor/backend/service/dm_service/internal/repository/postgres"
	"github.com/sudo-odner/minor/backend/service/dm_service/internal/transport/nats/producer"
)

type App struct {
	log *slog.Logger

	postgrespool *pgxpool.Pool
	nats         *nats.Conn
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
	repository := repo.NewChannelRepository(pool)

	// Init Nats
	nc, err := nats.Connect(
		cfg.Nats.URL,
		nats.Name("dm_service"),
		nats.Timeout(cfg.Nats.Timeout),
		nats.MaxReconnects(cfg.Nats.MaxReconnects),
		nats.ReconnectWait(cfg.Nats.ReconnectWait),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to NATS Core:%w", op, err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to initilize JetStream: %w", op, err)
	}
	a.nats = nc
	_ = producer.New(nc, js)

	// Init service

	// Init GRPC handler & service

	// Init HTTP handler & service

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
