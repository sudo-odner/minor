package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	events "github.com/sudo-odner/minor-shared/pkg/nats/events"
	eventsMessage "github.com/sudo-odner/minor-shared/pkg/nats/events/message"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/models"
)

type Producer struct {
	nc *nats.Conn
	js jetstream.JetStream
}

func NewProducer(nc *nats.Conn, js jetstream.JetStream) *Producer {
	return &Producer{
		nc: nc,
		js: js,
	}
}

// PublishMessageCreated publish event a message is created
func (p *Producer) PublishMessageCreated(ctx context.Context, msg models.Message) error {
	const op = "broker.nuts.PublishMessageCreated"

	event := eventsMessage.MessageCreatedEvent{
		MessageID: msg.MessageID,
		ChannelID: msg.ChannelID,
		UserID:    msg.UserID,
		Content:   msg.Content,
		ReplyTo:   msg.ReplyTo,
		CreatedAt: msg.CreatedAt,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := p.nc.Publish(events.SubjectMessageCreated, data); err != nil {
		return fmt.Errorf("%s: message not publish: %w", op, err)
	}

	return nil
}

// PublishMessageDeleted publish event a message is deleted
func (p *Producer) PublishMessageDeleted(ctx context.Context, channelID, messageID uuid.UUID) error {
	const op = "broker.nuts.PublishMessageDeleted"

	event := eventsMessage.MessageDeletedEvent{
		MessageID: messageID,
		ChannelID: channelID,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := p.nc.Publish(events.SubjectMessageDeleted, data); err != nil {
		return fmt.Errorf("%s: failed publish message: %w", op, err)
	}

	return nil
}
