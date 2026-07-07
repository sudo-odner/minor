package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/config"
	"go.uber.org/zap"
)

type App struct {
	Conn *nats.Conn
	JS   jetstream.JetStream
	log  *zap.Logger
}

func New(log *zap.Logger, cfg *config.Config) (*App, error) {
	opts := []nats.Option{
		nats.Name("auth_service"),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2 * time.Second),
	}

	nc, err := nats.Connect(cfg.NATS.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("jetstream init: %w", err)
	}

	return &App{
		Conn: nc,
		JS:   js,
		log:  log,
	}, nil
}

func (a *App) InitAuthStream(ctx context.Context) error {
	streamName := "AUTH_STREAM"
	
	cfg := jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{"auth.user.>", "auth.password.>"}, 
		Storage:  jetstream.FileStorage, 
		Retention: jetstream.LimitsPolicy,
		MaxAge:   7 * 24 * time.Hour,
	}

	_, err := a.JS.CreateOrUpdateStream(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create/update stream %s: %w", streamName, err)
	}

	a.log.Info("NATS JetStream initialized", zap.String("stream", streamName))
	return nil
}

func (a *App) Stop() {
	a.log.Info("closing nats connection")
	a.Conn.Close()
}