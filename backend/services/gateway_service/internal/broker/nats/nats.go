package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/service/gateway"
	"go.uber.org/zap"
)

type App struct {
	JS   jetstream.JetStream
	Conn *nats.Conn
	log  *zap.Logger
	hub *gateway.Hub
	// nodeID string // unique instance ID
}

func New(log *zap.Logger, cfg *config.Config) (*App, error) {
	opts := []nats.Option{
		nats.Name("gateway_service"),
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
		JS: js,
		Conn: nc,
		log: log,
	}, nil
}
