package app

import (
	"context"
	"fmt"

	grpcServ "github.com/sudo-odner/minor/backend/services/community_service/internal/app/grpc"
	httpServ "github.com/sudo-odner/minor/backend/services/community_service/internal/app/http"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/broker/nuts"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/repository/postgres"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/server/grpc"
	httpRouter "github.com/sudo-odner/minor/backend/services/community_service/internal/server/http"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/server/http/handler"
	presenceClient "github.com/sudo-odner/minor/backend/services/community_service/internal/client/grpc/presence"
	userClient "github.com/sudo-odner/minor/backend/services/community_service/internal/client/grpc/user"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/service/channel"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/service/members"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/service/permissions"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/service/roles"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/service/servers"
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

	// 1. Init repository Postgres
	repo, err := postgres.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// 2. Init broker Nuts
	broker, err := nuts.New(&cfg.Nuts)
	if err != nil {
		_ = repo.Close(ctx)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// grpc clients
	userAddr := cfg.GRPCClients.UserAddress // например "user-service:50051"
	userGRPCClient, err := userClient.NewUserClient(userAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to init user client: %w", err)
	}

	presenceAddr := cfg.GRPCClients.PresenceAddress // например "presence-service:50051"
	presenceGRPCClient, err := presenceClient.NewPresenceClient(presenceAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to init presence client: %w", err)
	}

	// 3. Init services
	permService := permissions.New(log, repo)
	serversService := servers.New(log, repo)
	rolesService := roles.New(log, repo, permService, serversService)
	channelService := channel.New(log, repo, broker, permService)
	membersService := members.New(log, repo, permService, serversService, userGRPCClient, presenceGRPCClient)

	// 4. Init HTTP handlers
	serverHandler := handler.NewServerHandler(log, serversService)
	channelHandler := handler.NewChannelHandler(log, channelService)
	memberHandler := handler.NewMemberHandler(log, membersService)
	roleHandler := handler.NewRoleHandler(log, rolesService)

	// 5. Init HTTP router and server
	router := httpRouter.NewRouter(log, httpRouter.Handlers{
		Server:  *serverHandler,
		Channel: *channelHandler,
		Member:  *memberHandler,
		Role:    *roleHandler,
	})
	httpServer := httpServ.New(&cfg.ServerHTTP, log, router)

	// 6. Init gRPC server
	gRPCHandler := grpc.New(log, permService)
	gRPCServer := grpcServ.New(&cfg.ServerGRPC, log, gRPCHandler)

	return &App{
		cfg:        cfg,
		log:        log,
		reposipory: repo,
		broker:     broker,
		grpcServer: gRPCServer,
		httpServer: httpServer,
	}, nil
}

func (a *App) Run() {
	const op = "app.Run"

	// Start gRPC server
	go func() {
		if err := a.grpcServer.Run(); err != nil {
			a.log.Error("failed to run gRPC server", zap.String("op", op), zap.Error(err))
		}
	}()

	// Start HTTP server
	if a.httpServer != nil {
		go func() {
			if err := a.httpServer.Run(); err != nil {
				a.log.Error("failed to run HTTP server", zap.String("op", op), zap.Error(err))
			}
		}()
	}
}

func (a *App) Stop(ctx context.Context) {
	const op = "app.Stop"

	// Stop HTTP server
	if a.httpServer != nil {
		if err := a.httpServer.Stop(ctx); err != nil {
			a.log.Error("failed to stop HTTP server", zap.String("op", op), zap.Error(err))
		}
	}

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

	// Close repository connections
	if a.reposipory != nil {
		_ = a.reposipory.Close(ctx)
	}
}
