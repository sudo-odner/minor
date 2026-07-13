package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/nats/events"
	communityEvents "github.com/sudo-odner/minor-shared/pkg/nats/events/community"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
)

func (b *Broker) PublishChannelCreated(ctx context.Context, serverID uuid.UUID, channel *models.Channel) error {
	const op = "broker.nuts.PublishChannelCreated"

	if err := b.publish(events.SubjectChannelCreated, communityEvents.ChannelEvent{
		ServerID: serverID,
		Channel:  toChannelDTO(channel),
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (b *Broker) PublishChannelUpdated(ctx context.Context, serverID uuid.UUID, channel *models.Channel) error {
	const op = "broker.nuts.PublishChannelUpdated"

	if err := b.publish(events.SubjectChannelUpdated, communityEvents.ChannelEvent{
		ServerID: serverID,
		Channel:  toChannelDTO(channel),
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (b *Broker) PublishChannelDeleted(ctx context.Context, serverID, channelID uuid.UUID) error {
	const op = "broker.nuts.PublishChannelDeleted"

	if err := b.publish(events.SubjectChannelDeleted, communityEvents.ChannelDeletedEvent{
		ServerID:  serverID,
		ChannelID: channelID,
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (b *Broker) PublishChannelPositionsUpdated(
	ctx context.Context,
	serverID uuid.UUID,
	channels []models.Channel,
) error {
	const op = "broker.nuts.PublishChannelPositionsUpdated"

	layouts := make([]communityEvents.ShortChannelDTO, len(channels))
	for i, ch := range channels {
		layouts[i] = toShortChannelDTO(ch)
	}

	event := communityEvents.ChannelPositionUpdateEvent{
		ServerID: serverID,
		Channels: layouts,
	}
	if err := b.publish(events.SubjectChannelPositionsUpdated, event); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// func (b *Broker) PublishMemberAdded(ctx context.Context, serverID string, user *models.Server) {
// 	const op = "broker.nuts.PublishMemberAdded"

// 	event := events.MemberAddedEvent{
// 		ServerID: serverID,
// 		Channels: layouts,
// 	}
// 	if err := b.publish(events.SubjectChannelPositionsUpdated, event); err != nil {
// 		return fmt.Errorf("%s: %w", op, err)
// 	}

// 	return nil
// }

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
