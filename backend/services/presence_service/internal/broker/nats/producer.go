package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sudo-odner/minor-shared/pkg/nats/events"
	presenceEvents	"github.com/sudo-odner/minor-shared/pkg/nats/events/presence"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/models"
)

func (b *Broker) PublishPresenceStatusUpdated(ctx context.Context, p *models.Presence) error {
	const op = "broker.nuts.PublishPresenceStatusUpdated"

	event := presenceEvents.PresenceStatusUpdatedEvent{
		UserID:       p.UserID,
		Status:       int32(p.Status),
		CustomStatus: p.CustomStatus,
		LastActiveAt: p.LastActiveAt,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("%s: marshal failed: %w", op, err)
	}

	err = b.conn.Publish(events.SubjectPresenceStatusUpdated, data)
	if err != nil {
		return fmt.Errorf("%s: publish failed: %w", op, err)
	}

	return nil
}
