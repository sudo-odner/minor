package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/config"
	"go.uber.org/zap"
)

type App struct {
	log        *zap.Logger
	httpServer *http.Server
}

func New(log *zap.Logger, cfg *config.Config) *App {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	httpServer := http.Server{
		Addr:         cfg.ServerConfig.Port,
		Handler:      mux,
		ReadTimeout:  cfg.ServerConfig.Timeout,
		WriteTimeout: cfg.ServerConfig.Timeout,
		IdleTimeout:  cfg.ServerConfig.IdleTimeout,
	}

	return &App{log: log, httpServer: &httpServer}
}

func (a *App) Run() error {
	const op = "httpapp.Run"

	log := a.log.With(
		zap.String("op", op),
		zap.String("port", a.httpServer.Addr),
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
		zap.String("op", op),
		zap.String("port", a.httpServer.Addr),
	)

	log.Info("stoping http server")

	if err := a.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("server stoped")

	return nil
}
