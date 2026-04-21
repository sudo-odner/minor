package http_serv

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sudo-odner/minor/backend/services/chat_service/internal/config"
)

type HttpServ struct {
	server *http.Server
}

func New(cfg *config.HttpServer, handler http.Handler) *HttpServ {
	return &HttpServ{
		server: &http.Server{
			Addr:        cfg.Address,
			Handler:     handler,
			ReadTimeout: cfg.Timeout,
			IdleTimeout: cfg.IdleTimeout,
		},
	}
}

func (hs *HttpServ) Address() string {
	return hs.server.Addr
}

func (hs *HttpServ) Run() error {
	const op = "app.httpServ.Run"
	if err := hs.server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (hs *HttpServ) Stop(ctx context.Context) error {
	const op = "app.httpServ.Stop"
	if err := hs.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
