package notifier

import (
	"context"
	"fmt"

	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/client/grpc/presence"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/models"
	"go.uber.org/zap"
)

type PushProvider interface {
	Send(ctx context.Context, userID string, title, body string) error
}

type Notifier struct {
	log *zap.Logger
	presenceClient *presence.Client
	pushProvider PushProvider
}

func NewNotifier(presenceClient *presence.Client, pushProvider PushProvider) *Notifier {
	return &Notifier{
		presenceClient: presenceClient,
		pushProvider: pushProvider,
	}
}

func (n *Notifier) HandlerChatMessage(ctx context.Context, event models.ChatMessageCreated) error {
	log := n.log.With(
		zap.String("op", "notifier"),
	)
	
	isOnline, err := n.presenceClient.IsUserOnline(ctx, event.AuthorID)
	if err != nil {
		return fmt.Errorf("failed to check presence: %w", err)
	}

	if isOnline {
		log.Info("user is online, skipping push", zap.String("user", event.AuthorID))
		return nil
	}

	title := "New Message"
	body := fmt.Sprintf("User %s: %s", event.AuthorID, truncate(event.Content, 50))

	err = n.pushProvider.Send(ctx, event.AuthorID, title, body)
	if err != nil {
		return fmt.Errorf("push delivery failed: %w", err)
	}

	log.Info("push sent successfully", zap.String("user", event.AuthorID))
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}

	return s[:max] + "..."
}