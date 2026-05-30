package app

import (
	"context"
	"fmt"

	grpcServ "github.com/sudo-odner/minor/backend/services/user_service/internal/app/grpc"
	httpServer "github.com/sudo-odner/minor/backend/services/user_service/internal/app/http"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/broker/nuts"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/repository/postgres"
	grpcRouter "github.com/sudo-odner/minor/backend/services/user_service/internal/server/grpc"
	httpRouter "github.com/sudo-odner/minor/backend/services/user_service/internal/server/http"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/server/http/handler"
	friendService "github.com/sudo-odner/minor/backend/services/user_service/internal/service/friend"
	userService "github.com/sudo-odner/minor/backend/services/user_service/internal/service/user"
	"go.uber.org/zap"
)

type App struct {
	log        *zap.Logger
	repository *postgres.Repository
	broker     *nuts.Broker
	httpServer *httpServer.Server
	grpcServer *grpcServ.Server
	ErrChan    chan error
}

func New(cfg *config.Config, log *zap.Logger) (*App, error) {
	const op = "app.New"
	ctx := context.Background()

	// 1. Initialize PostgreSQL Repository
	storageDSN := cfg.Postgres.DSN()
	repo, err := postgres.New(ctx, storageDSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// 2. Initialize Nats Broker
	broker, err := nuts.New(&cfg.Nuts)
	if err != nil {
		_ = repo.Close(ctx)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// 3. Initialize Services
	usrService := userService.New(log, repo, broker)
	frnService := friendService.New(log, repo, broker)

	// 4. Initialize HTTP Handlers
	uHandler := handler.NewUserHandler(log, usrService)
	fHandler := handler.NewFriendHandler(log, frnService)

	// 5. Initialize HTTP Router
	router := httpRouter.NewRouter(log, httpRouter.Handlers{
		User:   uHandler,
		Friend: fHandler,
	})

	// 6. Initialize HTTP Server
	server := httpServer.New(&cfg.ServerHTTP, log, router)

	// 7. Initialize gRPC Handler and Server
	gRPCHandler := grpcRouter.New(log, usrService)
	gRPCServer := grpcServ.New(&cfg.ServerGRPC, log, gRPCHandler)

	return &App{
		log:        log,
		repository: repo,
		broker:     broker,
		httpServer: server,
		grpcServer: gRPCServer,
		ErrChan:    make(chan error, 1),
	}, nil
}

func (a *App) Run() {
	const op = "app.Run"

	// Start gRPC server in background
	go func() {
		if err := a.grpcServer.Run(); err != nil {
			a.log.Error("failed to run gRPC server", zap.String("op", op), zap.Error(err))
		}
	}()

	// Start HTTP server
	if err := a.httpServer.Run(); err != nil {
		a.ErrChan <- err
	}
}

func (a *App) Stop(ctx context.Context) error {
	const op = "app.Stop"
	a.log.Info("stopping application")

	var stopErr error
	if a.httpServer != nil {
		if err := a.httpServer.Stop(ctx); err != nil {
			a.log.Error("failed to stop HTTP server", zap.Error(err))
			stopErr = err
		}
	}

	if a.grpcServer != nil {
		if err := a.grpcServer.Stop(ctx); err != nil {
			a.log.Error("failed to stop gRPC server", zap.Error(err))
			if stopErr == nil {
				stopErr = err
			}
		}
	}

	if a.broker != nil {
		if err := a.broker.Close(ctx); err != nil {
			a.log.Error("failed to close Nats broker", zap.Error(err))
			if stopErr == nil {
				stopErr = err
			}
		}
	}

	if a.repository != nil {
		if err := a.repository.Close(ctx); err != nil {
			a.log.Error("failed to close database connection", zap.Error(err))
			if stopErr == nil {
				stopErr = err
			}
		}
	}

	return stopErr
}
