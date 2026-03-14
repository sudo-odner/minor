package app

import (
	"context"
	"net/http"

	httpServer "github.com/sudo-odner/minor/backend/services/user_service/internal/app/http"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/config"
	"go.uber.org/zap"
)

type Deps struct {
	Config *config.Config
	Logger *zap.Logger
}

type App struct {
	log        *zap.Logger
	httpServer *httpServer.Server
	ErrChan    chan error
}

func New(d Deps) *App {
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
		log:        d.Logger,
		httpServer: httpServer.New(d.Config, d.Logger, mux),
		ErrChan:    make(chan error, 1),
	}
}

func (a *App) Run() {
	if err := a.httpServer.Run(); err != nil {
		a.ErrChan <- err
	}
}

func (a *App) Stop(ctx context.Context) error {
	return a.httpServer.Stop(ctx)
}
