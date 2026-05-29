package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/models"
	"github.com/sudo-odner/minor/backend/shared/pkg/events"
)

type AuthPublisher struct {
	js jetstream.JetStream
}

func NewAuthPublisher(js jetstream.JetStream) *AuthPublisher {
	return &AuthPublisher{js: js}
}

func (p *AuthPublisher) PublishLoginSuccess(ctx context.Context, userID, ip, userAgent string) error {
	event := events.UserLoginSuccessEvent{
		UserID:    userID,
		Timestamp: time.Now(),
		IP:        ip,
		UserAgent: userAgent,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal login_success event: %w", err)
	}

	_, err = p.js.Publish(ctx, "auth.user.login_success", data)
	if err != nil {
		return fmt.Errorf("failed to publish login event: %w", err)
	}

	return nil
}

func (p *AuthPublisher) PublishUserRegistered(ctx context.Context, user *models.User) error {
	event := events.UserRegisteredEvent{
		UserID:    user.ID.String(),
		Timestamp: time.Now(),
		Email:     user.Email,
	}

	data, _ := json.Marshal(event)

	_, err := p.js.Publish(ctx, "auth.user.login_success", data)
	if err != nil {
		return fmt.Errorf("failed to publish login event: %w", err)
	}

	return nil
}

func (p *AuthPublisher) PublishUserLoggedOut(ctx context.Context, userID, tokenID string) error {
	event := events.UserLoggedOutEvent{
		UserID:    userID,
		TokenID:   tokenID,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal logout event: %w", err)
	}

	_, err = p.js.Publish(ctx, "auth.user.logged_out", data)
	if err != nil {
		return fmt.Errorf("failed to publish logout event to NATS: %w", err)
	}

	return nil
}
