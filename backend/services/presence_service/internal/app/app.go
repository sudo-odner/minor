package app

import (
	"context"
	"fmt"

	grpcServ "github.com/sudo-odner/minor/backend/services/presence_service/internal/app/grpc"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/broker/nuts"

	"github.com/sudo-odner/minor/backend/services/presence_service/internal/cache/redis"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/server/grpc"
	presenceService "github.com/sudo-odner/minor/backend/services/presence_service/internal/service/presence"
	"go.uber.org/zap"
)

type App struct {
	cfg        *config.Config
	log        *zap.Logger
	cache      *redis.Cache
	broker     *nuts.Broker
	grpcServer *grpcServ.Server
}

func New(cfg *config.Config, log *zap.Logger) (*App, error) {
	const op = "app.New"
	ctx := context.Background()

	// 1. Init cache Redis
	cache, err := redis.New(ctx, cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// 2. Init broker Nuts
	broker, err := nuts.New(&cfg.Nuts)
	if err != nil {
		_ = cache.Stop()
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// 3. Init services
	presenceS := presenceService.New(log, cache, broker)

	// 4. Init gRPC server
	gRPCHandler := grpc.New(log, presenceS)
	gRPCServer := grpcServ.New(&cfg.ServerGRPC, log, gRPCHandler)

	return &App{
		cfg:        cfg,
		log:        log,
		cache:      cache,
		broker:     broker,
		grpcServer: gRPCServer,
	}, nil
}

func (a *App) Run() {
	const op = "app.Run"

	// Start gRPC server
	if err := a.grpcServer.Run(); err != nil {
		a.log.Error("failed to run gRPC server", zap.String("op", op), zap.Error(err))
	}
}

func (a *App) Stop(ctx context.Context) {
	const op = "app.Stop"

	// Stop gRPC server
	if a.grpcServer != nil {
		if err := a.grpcServer.Stop(ctx); err != nil {
			a.log.Error("failed to stop gRPC server", zap.String("op", op), zap.Error(err))
		}
	}

	// Close broker connection
	if a.broker != nil {
		if err := a.broker.Close(ctx); err != nil {
			a.log.Error("failed to close broker", zap.String("op", op), zap.Error(err))
		}
	}

	// Close cahce connections
	if a.cache != nil {
		if err := a.cache.Stop(); err != nil {
			a.log.Error("failed to stop cache", zap.String("op", op), zap.Error(err))
		}
	}
}
