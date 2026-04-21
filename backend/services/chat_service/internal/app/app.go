package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpServ "github.com/sudo-odner/minor/backend/services/chat_service/internal/app/http_serv"
	"github.com/sudo-odner/minor/backend/services/chat_service/internal/config"
	"go.uber.org/zap"
)

type App struct {
	log      *zap.Logger
	httpServ *httpServ.HttpServ
}

func New(cfg *config.Config, log *zap.Logger) (*App, error) {
	// TODO: repo cassandra
	// TODO: brocker nuts
	// TODO: services
	// TODO: handler

	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})
	})

	return &App{
		log:      log,
		httpServ: httpServ.New(&cfg.HttpServer, router),
	}, nil
}

func (a *App) Run() error {
	const op = "app.Run"
	log := a.log.With(zap.String("op", op))

	log.Info("starting application")

	log.Info("starting http server", zap.String("address", a.httpServ.Address()))
	if err := a.httpServ.Run(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	const op = "app.Stop"
	log := a.log.With(zap.String("op", op))

	log.Info("stopping application")
	if err := a.httpServ.Stop(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("http server stopped")

	log.Info("application stopped successfully")
	return nil
}
