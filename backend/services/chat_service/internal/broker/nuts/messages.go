package nuts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/chat_service/internal/models"
)

func (b *Broker) PublishMessageCreated(ctx context.Context, msg *models.Message) error {
	const op = "broker.nuts.PublishMessageCreated"

	event := MessageCreatedEvent{
		MessageID: msg.MessageID.String(),
		ChannelID: msg.ChannelID.String(),
		AuthorID:  msg.AuthorID.String(),
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt,
	}
	if msg.ReplyTo != nil {
		replyTo := msg.ReplyTo.String()
		event.ReplyTo = &replyTo
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := b.conn.Publish(SubjectMessageCreated, data); err != nil {
		return fmt.Errorf("%s: message not publish: %w", op, err)
	}

	return nil
}

func (b *Broker) PublishMessageDeleted(ctx context.Context, channelID, messageID uuid.UUID) error {
	return nil
}
