package nuts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
)

func (b *Broker) PublishChannelCreated(ctx context.Context, serverID uuid.UUID, channel *models.Channel) error {
	return nil
}

func (b *Broker) PublishChannelUpdated(ctx context.Context, serverID uuid.UUID, channel *models.Channel) error {
	return nil
}

func (b *Broker) PublishChannelDeleted(ctx context.Context, serverID uuid.UUID, channel *models.Channel) error {
	return nil
}

func (b *Broker) PubllisChannelPositionsUpdated(
	ctx context.Context,
	serverID uuid.UUID,
	parentID *uuid.UUID,
	channel *models.Channel,
) error {
	return nil
}

func (b *Broker) publish(subject string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal falied for subject %s: %w", subject, err)
	}

	if err := b.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("publish to nats failed for subject %s: %w", subject, err)
	}

	return nil
}
