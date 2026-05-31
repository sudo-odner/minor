package notifier

import (
	"context"
	"fmt"

	"github.com/sudo-odner/minor/backend/services/notification_service/internal/client/grpc/presence"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/models"
)

// EmailProvider — наш новый интерфейс вместо Push
type EmailProvider interface {
	Send(ctx context.Context, toEmail string, title, body string) error
}

// UserClient — интерфейс gRPC клиента к User Service
type UserClient interface {
	GetUserEmail(ctx context.Context, userID string) (string, error)
	GetUserName(ctx context.Context, userID string) (string, error)
}

type Notifier struct {
	presenceClient *presence.Client // из прошлого примера
	userClient     UserClient       // Добавили связь с User Service
	emailProvider  EmailProvider
}

func NewNotifier(presenceClient *presence.Client, userClient UserClient, emailProvider EmailProvider) *Notifier {
	return &Notifier{
		presenceClient: presenceClient,
		userClient:     userClient,
		emailProvider:  emailProvider,
	}
}

func (n *Notifier) HandleChatMessage(ctx context.Context, event models.ChatMessageCreated) error {
	isOnline, _ := n.presenceClient.IsUserOnline(ctx, event.AuthorID)
	if isOnline {
		return nil
	}

	email, err := n.userClient.GetUserEmail(ctx, event.AuthorID)
	if err != nil {
		return err
	}

	// Используем твою логику текста, но через интерфейс
	title := "Оповещение о сообщении в Minor"
	body := fmt.Sprintf("Здравствуйте!\n\nВ канале появилось новое сообщение:\n%s", event.Content)

	return n.emailProvider.Send(ctx, email, title, body)
}

// Добавляем обработку события входа (если прилетит из NATS)
func (n *Notifier) HandleLoginEvent(ctx context.Context, userID, ip string) error {
    email, _ := n.userClient.GetUserEmail(ctx, userID)
    userName, _ := n.userClient.GetUserName(ctx, userID)
    
    title := "Оповещение о входе в Minor"
    body := fmt.Sprintf("Здравствуйте, %s!\n\nБыл выполнен вход в ваш аккаунт.\nIP-адрес: %s", userName, ip)
    
    return n.emailProvider.Send(ctx, email, title, body)
}
