package httpserv

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sudo-odner/minor/backend/services/message_service/internal/config"
)

type HTTPServ struct {
	server *http.Server
}

func New(cfg *config.HTTPServer, handler http.Handler) *HTTPServ {
	return &HTTPServ{
		server: &http.Server{
			Addr:        cfg.Address,
			Handler:     handler,
			ReadTimeout: cfg.Timeout,
			IdleTimeout: cfg.IdleTimeout,
		},
	}
}

func (hs *HTTPServ) Address() string {
	return hs.server.Addr
}

func (hs *HTTPServ) Run() error {
	const op = "app.httpServ.Run"
	if err := hs.server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (hs *HTTPServ) Stop(ctx context.Context) error {
	const op = "app.httpServ.Stop"
	if err := hs.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
