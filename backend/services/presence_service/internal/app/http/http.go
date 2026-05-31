package httpServ

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sudo-odner/minor/backend/services/community_service/internal/config"
	"go.uber.org/zap"
)

type Server struct {
	cfg    *config.ServerHTTP
	log    *zap.Logger
	server *http.Server
}

func New(cfg *config.ServerHTTP, log *zap.Logger, handler http.Handler) *Server {
	return &Server{
		cfg: cfg,
		log: log,
		server: &http.Server{
			Addr:        cfg.Address,
			ReadTimeout: cfg.Timeout,
			IdleTimeout: cfg.IdleTimeout,
			Handler:     handler,
		},
	}
}

func (s *Server) Run() error {
	const op = "app.http.Run"
	s.log.Info("starting HTTP server", zap.String("addr", s.cfg.Address))

	if err := s.server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	const op = "app.http.Stop"
	s.log.Info("stopping HTTP server", zap.String("addr", s.cfg.Address))

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
