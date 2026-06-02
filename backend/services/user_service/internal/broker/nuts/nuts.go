package nuts

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream" // Используем новый пакет для JS
	"github.com/sudo-odner/minor/backend/services/user_service/internal/config"
	"go.uber.org/zap"
)

type Broker struct {
	log  *zap.Logger
	conn *nats.Conn
	JS   jetstream.JetStream 
}

func New(cfg *config.Nuts, log *zap.Logger) (*Broker, error) {
	const op = "broker.nuts.New"

	opts := []nats.Option{
		nats.Name("user_service"), // Имя клиента в сети NATS
		nats.Timeout(cfg.Timeout),
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.TimeoutReconnect),
	}

	// 1. Устанавливаем базовое соединение
	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("%s: connect to nats failed: %w", op, err)
	}

	// 2. Инициализируем подсистему JetStream
	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("%s: failed to setup jetstream: %w", op, err)
	}

	log.Info("nats jetstream initialized", zap.String("url", cfg.URL))

	return &Broker{
		log:  log,
		conn: conn,
		JS:   js,
	}, nil
}

// Ping проверяет состояние соединения
func (b *Broker) Ping(ctx context.Context) error {
	if b.conn == nil || !b.conn.IsConnected() {
		return fmt.Errorf("nats connection is lost")
	}
	return nil
}

// Close корректно закрывает соединение
func (b *Broker) Close(ctx context.Context) error {
	if b.conn != nil {
		b.log.Info("closing nats connection")
		b.conn.Close()
	}
	return nil
}