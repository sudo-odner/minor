package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/sudo-odner/min/backend/auth-service/internal/config"
)

type App struct {
	log *slog.Logger
	httpServer *http.Server
}

func New(log *slog.Logger, cfg *config.Config, router chi.Router) *App {
	httpServer := http.Server{
		Addr: cfg.ServerConfig.Port,
		Handler: router,
		ReadTimeout: cfg.ServerConfig.Timeout,
		WriteTimeout: cfg.ServerConfig.Timeout,
		IdleTimeout: cfg.ServerConfig.IdleTimeout,
	}

	return &App{log: log, httpServer: &httpServer}
}

func (a *App) Run() error {
	const op = "httpapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.String("port", a.httpServer.Addr),
	)

	log.Info("starting http server")

	if err := a.httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("http server started")

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	const op = "httpapp.Stop"

	log := a.log.With(
		slog.String("op", op),
		slog.String("port", a.httpServer.Addr),
	)

	log.Info("stoping http server")

	if err := a.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("server stoped")

	return nil
}