package nuts

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/config"
)

type Broker struct {
	conn *nats.Conn
}

func New(cfg *config.Nuts) (*Broker, error) {
	const op = "broker.nuts.New"

	opts := []nats.Option{
		nats.Name("community_service"),
		nats.Timeout(cfg.Timeout),
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.TimeoutReconnect),
	}

	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("%s: connect to nuts failed: %w", op, err)
	}

	return &Broker{
		conn: conn,
	}, nil
}

func (b *Broker) Ping(ctx context.Context) error {
	if b.conn == nil || !b.conn.IsConnected() {
		return fmt.Errorf("nats connection is lost")
	}
	return nil
}

func (b *Broker) Close(ctx context.Context) error {
	if b.conn != nil {
		b.conn.Close()
	}
	return nil
}
