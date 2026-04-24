package nuts

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/sudo-odner/minor/backend/services/chat_service/internal/config"
)

type Broker struct {
	conn *nats.Conn
}

func New(cfg config.Nuts, url string) (*Broker, error) {
	const op = "broker.nuts.New"

	nc, err := nats.Connect(url,
		nats.Name("caht_service"),
		nats.Timeout(cfg.Timeout),
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.ReconnectWait),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Broker{
		conn: nc,
	}, nil
}

func (b *Broker) Ping() error {
	if b.conn == nil || !b.conn.IsConnected() {
		return fmt.Errorf("nats connection is lost")
	}
	return nil
}

func (b *Broker) Stop() error {
	if b.conn != nil {
		b.conn.Close()
	}
	return nil
}
