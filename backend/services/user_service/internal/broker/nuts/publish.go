package nuts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/models"
)

func (b *Broker) PublishUserCreated(ctx context.Context, u *models.User) error {
	const op = "broker.nuts.PublishUserCreated"

	event := UserCreatedEvent{
		UserID:   u.ID.String(),
		Username: u.Username, Email: u.Email,
	}

	if err := b.publish(SubjectUserCreated, event); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (b *Broker) PublishUserUpdated(ctx context.Context, u *models.User) error {
	const op = "broker.nuts.PublishUserUpdated"

	event := UserUpdatedEvent{
		UserID:    u.ID.String(),
		Username:  u.Username,
		AvatarURL: u.AvatarURL,
		Bio:       u.Bio,
	}

	if err := b.publish(SubjectUserUpdated, event); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (b *Broker) PublishUserDeleted(ctx context.Context, userID uuid.UUID) error {
	const op = "broker.nuts.PublishUserDeleted"

	if err := b.publish(SubjectUserDeleted, UserDeletedEvent{
		UserID: userID.String(),
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (b *Broker) PublishRelationshipUpdated(ctx context.Context, userID, targetID uuid.UUID, status models.RelationshipStatus) error {
	const op = "broker.nuts.PublishRelationshipUpdated"

	event := RelationshipUpdatedEvent{
		UserID:   userID.String(),
		TargetID: targetID.String(),
		Status:   int16(status),
	}

	if err := b.publish(SubjectRelationshipUpdated, event); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (b *Broker) PublishRelationshipDeleted(ctx context.Context, userID, targetID uuid.UUID) error {
	const op = "broker.nuts.PublishRelationshipDeleted"

	event := RelationshipDeletedEvent{
		UserID:   userID.String(),
		TargetID: targetID.String(),
	}

	if err := b.publish(SubjectRelationshipDeleted, event); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
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
