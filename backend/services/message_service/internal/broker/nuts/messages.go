package nuts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/events"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/models"
)

func (b *Broker) PublishMessageCreated(ctx context.Context, msg models.Message) error {
	const op = "broker.nuts.PublishMessageCreated"

	event := events.MessageCreatedEvent{
		MessageID: msg.MessageID,
		ChannelID: msg.ChannelID,
		AuthorID:  msg.UserID,
		Username: msg.Username,
		Content:   msg.Content,
		ReplyTo:   msg.ReplyTo,
		CreatedAt: msg.CreatedAt,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := b.conn.Publish(events.SubjectMessageCreated, data); err != nil {
		return fmt.Errorf("%s: message not publish: %w", op, err)
	}

	return nil
}

func (b *Broker) PublishMessageDeleted(ctx context.Context, channelID, messageID uuid.UUID) error {
	const op = "broker.nuts.PublishMessageDeleted"

	event := events.MessageDeletedEvent{
		MessageID: messageID,
		ChannelID: channelID,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := b.conn.Publish(events.SubjectMessageDeleted, data); err != nil {
		return fmt.Errorf("%s: failed publish message: %w", op, err)
	}

	return nil
}
