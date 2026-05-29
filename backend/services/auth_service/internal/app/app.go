package app

import (
	"github.com/go-chi/chi/v5"
	authv1 "github.com/sudo-odner/minor-shared/pkg/pb/auth/v1"
	grpcapp "github.com/sudo-odner/minor/backend/services/auth_service/internal/app/grpc"
	httpapp "github.com/sudo-odner/minor/backend/services/auth_service/internal/app/http"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/config"
	"go.uber.org/zap"
)

type App struct {
	HTTPServer *httpapp.App
	GRPCServer *grpcapp.App
	log        *zap.Logger
}

func New(log *zap.Logger, cfg *config.Config, router chi.Router, authService authv1.AuthServiceServer ) *App {
	httpApp := httpapp.New(log, cfg, router)
	grpcApp := grpcapp.New(log, authService, cfg.GRPC.Port)
	
	return &App{
		HTTPServer: httpApp,
		GRPCServer: grpcApp,
		log:        log,
	}
}
