package app

import (
	"context"
	"fmt"

	grpcServ "github.com/sudo-odner/minor/backend/services/community_service/internal/app/grpc"
	httpServ "github.com/sudo-odner/minor/backend/services/community_service/internal/app/http"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/broker/nuts"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/repository/postgres"
	"go.uber.org/zap"
)

type App struct {
	cfg        *config.Config
	log        *zap.Logger
	reposipory *postgres.Repository
	broker     *nuts.Broker
	httpServer *httpServ.Server
	grpcServer *grpcServ.Server
}

func New(cfg *config.Config, log *zap.Logger) (*App, error) {
	const op = "app.New"
	ctx := context.Background()

	// Init repository Postgres
	repo, err := postgres.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	// TODO: Init broker Nuts
	// TODO: Init service
	// TODO: Init HTTP server
	// TODO: Init gRPC server
	return &App{
		cfg: cfg,
		log: log,
	}, nil
}

func (a *App) Run() {
}

func (a *App) Stop(ctx context.Context) {
}
