package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/config"
	"go.uber.org/zap"
)

type GatewayHTTPServer struct {
	log        *zap.Logger
	httpServer *http.Server
}

func New(log *zap.Logger, cfg *config.Config, router chi.Router) *GatewayHTTPServer {
	httpServer := http.Server{
		Addr:         cfg.HTTP.Port,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.Timeout,
		WriteTimeout: cfg.HTTP.Timeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return &GatewayHTTPServer{log: log, httpServer: &httpServer}
}
	
func (gs *GatewayHTTPServer) Run() error {
	const op = "app.http.Run"

	log := gs.log.With(
		zap.String("op", op),
	)

	log.Info("starting http server")

	if err := gs.httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("http server started")

	return nil
}

func (gs *GatewayHTTPServer) Stop(ctx context.Context) error {
	const op = "app.http.Stop"
	
	log := gs.log.With(
		zap.String("op", op),
	)

	log.Info("trying to gracefully stop gateway http server")
	
	if err := gs.httpServer.Shutdown(ctx); err != nil {
		log.Error("failed to gracefully stop gateway http server")
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("server stopped successfully")

	return nil
}
