package app

import (
	"context"
	"fmt"
	"net/http"

	httpServer "github.com/sudo-odner/minor/backend/services/user_service/internal/app/http"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/repository/postgres"
	"go.uber.org/zap"
)

type App struct {
	log        *zap.Logger
	httpServer *httpServer.Server
	ErrChan    chan error
}

func New(cfg *config.Config, log *zap.Logger) (*App, error) {
	const op = "app.New"

	storageDSN := cfg.PostgreConfig.DSN()
	_, err := postgres.New(context.Background(), storageDSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Initialize usecase
	// Initialize router
	// Заглушка для теста
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	// ------------------
	return &App{
		log:        log,
		httpServer: httpServer.New(cfg, log, mux),
		ErrChan:    make(chan error, 1),
	}, nil
}

func (a *App) Run() {
	if err := a.httpServer.Run(); err != nil {
		a.ErrChan <- err
	}
}

func (a *App) Stop(ctx context.Context) error {
	return a.httpServer.Stop(ctx)
}
