package nuts

import (
	"context"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/chat_service/internal/models"
)

func (b *Broker) PublishMessageCreated(ctx context.Context, msg *models.Message) error {
	return nil
}

func (b *Broker) PublishMessageDeleted(ctx context.Context, channelID, messageID uuid.UUID) error {
	return nil
}
