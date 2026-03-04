package app

import (
	"github.com/go-chi/chi/v5"
	httpapp "github.com/sudo-odner/min/backend/auth-service/internal/app/http"
	"github.com/sudo-odner/min/backend/auth-service/internal/config"
	"go.uber.org/zap"
)

type App struct {
	HTTPServer *httpapp.App
	log *zap.Logger
}

func New(log *zap.Logger, cfg *config.Config, router chi.Router) *App {
	httpApp := httpapp.New(log, cfg, router)

	return &App{
		HTTPServer: httpApp,
		log: log,
	}
}

