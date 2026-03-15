package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sudo-odner/minor/backend/services/user_service/internal/config"
	"go.uber.org/zap"
)

type Server struct {
	log        *zap.Logger
	httpServer *http.Server
}

func New(cfg *config.Config, log *zap.Logger, hander http.Handler) *Server {
	return &Server{
		log: log,
		httpServer: &http.Server{
			Addr:        cfg.ServerConfig.Host + ":" + cfg.ServerConfig.Port,
			Handler:     hander,
			ReadTimeout: cfg.ServerConfig.Timeout,
			IdleTimeout: cfg.ServerConfig.IdleTimeout,
		},
	}
}

func (s *Server) Run() error {
	const op = "app.http.Run"

	log := s.log.With(
		zap.String("op", op),
		zap.String("addr", s.httpServer.Addr),
	)

	log.Info("starting http server")

	if err := s.httpServer.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil

}
func (s *Server) Stop(ctx context.Context) error {
	const op = "app.http.Stop"

	log := s.log.With(
		zap.String("op", op),
		zap.String("addr", s.httpServer.Addr),
	)

	log.Info("stopping http server")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("stopped http server")

	return nil
}
