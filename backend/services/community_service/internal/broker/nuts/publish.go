package nuts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
)

// Хз поидее разные логики и если что то поменяется то будет окей. Но не совсем увререн в каком это враианте будет
// Поидее можно их объеденить Created и Updated
func (b *Broker) PublishChannelCreated(ctx context.Context, serverID uuid.UUID, channel *models.Channel) error {
	const op = "broker.nuts.PublishChannelCreated"

	if err := b.publish(SubjectChannelCreated, ChannelEvent{
		ServerID: serverID.String(),
		Channel:  toChannelDTO(channel),
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (b *Broker) PublishChannelUpdated(ctx context.Context, serverID uuid.UUID, channel *models.Channel) error {
	const op = "broker.nuts.PublishChannelUpdated"

	if err := b.publish(SubjectChannelUpdated, ChannelEvent{
		ServerID: serverID.String(),
		Channel:  toChannelDTO(channel),
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (b *Broker) PublishChannelDeleted(ctx context.Context, serverID, channelID uuid.UUID) error {
	const op = "broker.nuts.PublishChannelDeleted"

	if err := b.publish(SubjectChannelDeleted, ChannelDeletedEvent{
		ServerID:  serverID.String(),
		ChannelID: channelID.String(),
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

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
