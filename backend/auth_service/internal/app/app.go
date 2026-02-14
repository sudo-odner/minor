package app

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	httpapp "github.com/sudo-odner/min/backend/auth-service/internal/app/http"
	"github.com/sudo-odner/min/backend/auth-service/internal/config"
)

type App struct {
	HTTPServer *httpapp.App
	log *slog.Logger
}

func New(log *slog.Logger, cfg *config.Config, router chi.Router) *App {
	httpApp := httpapp.New(log, cfg, router)

	return &App{
		HTTPServer: httpApp,
		log: log,
	}
}

