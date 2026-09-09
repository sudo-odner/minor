package producer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sudo-odner/minor-shared/pkg/nats/events"
)

type Producer struct {
	nc *nats.Conn
	js jetstream.JetStream
}

func New(nc *nats.Conn, js jetstream.JetStream) *Producer {
	return &Producer{
		nc: nc,
		js: js,
	}
}

func (p *Producer) Publish(ctx context.Context, evt events.Event) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal event: %T: %w", evt, err)
	}

	subject := evt.Subject()
	if _, err = p.js.Publish(ctx, subject, data); err != nil {
		return fmt.Errorf("publish event %T to %s: %w", evt, subject, err)
	}
	return nil
}
