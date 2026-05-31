package nuts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sudo-odner/minor/backend/services/presence_service/internal/models"
)

type PresenceStatusUpdatedEvent struct {
	UserID       string            `json:"user_id"`
	Status       models.UserStatus `json:"status"`
	CustomStatus string            `json:"custom_status"`
	LastActiveAt int64             `json:"last_active_at"`
}

func (b *Broker) PublishPresenceStatusUpdated(ctx context.Context, p *models.Presence) error {
	const op = "broker.nuts.PublishPresenceStatusUpdated"

	event := PresenceStatusUpdatedEvent{
		UserID:       p.UserID,
		Status:       p.Status,
		CustomStatus: p.CustomStatus,
		LastActiveAt: p.LastActiveAt,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("%s: marshal failed: %w", op, err)
	}

	err = b.conn.Publish(SubjectPresenceStatusUpdated, data)
	if err != nil {
		return fmt.Errorf("%s: publish failed: %w", op, err)
	}

	return nil
}
