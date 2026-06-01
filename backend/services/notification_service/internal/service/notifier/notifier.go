package notifier

import (
	"context"
	"fmt"

	"github.com/sudo-odner/minor-shared/pkg/events"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/client/grpc/presence"
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

func (n *Notifier) HandleRegistration(ctx context.Context, event events.UserRegisteredEvent) error {
	title := "Добро пожаловать в Minor!"
	body := fmt.Sprintf("Здравствуйте, %s!\n\nВы успешно зарегистрировались. Ваш ID: %s", 
		event.Email, event.UserID)

	return n.emailProvider.Send(ctx, event.Email, title, body)
}

// HandleLogin отправляет уведомление о входе
func (n *Notifier) HandleLogin(ctx context.Context, event events.UserLoginSuccessEvent) error {
	// Сначала узнаем email и имя через gRPC, так как в событии Login их обычно нет (только ID)
	email, _ := n.userClient.GetUserEmail(ctx, event.UserID)
	username, _ := n.userClient.GetUserName(ctx, event.UserID)

	title := "Новый вход в аккаунт"
	body := fmt.Sprintf("Привет, %s!\n\nВ твой аккаунт вошли.\nIP: %s\nВремя: %s", 
		username, event.IP, event.Timestamp.Format("15:04:05 02.01.2006"))

	return n.emailProvider.Send(ctx, email, title, body)
}

func (n *Notifier) HandleChatMessage(ctx context.Context, event events.MessageCreatedEvent) error {
	// 1. Проверяем статус (Online/Offline)
	// Используем AuthorID или RecipientID из твоего события
	isOnline, err := n.presenceClient.IsUserOnline(ctx, event.AuthorID.String())
	if err != nil {
		return err
	}

	if isOnline {
		return nil
	}

	// 2. Получаем данные и шлем письмо
	email, err := n.userClient.GetUserEmail(ctx, event.AuthorID.String())
	if err != nil {
		return err
	}

	title := "Новое сообщение в Minor"
	body := fmt.Sprintf("Вам пришло сообщение: %s", event.Content)

	return n.emailProvider.Send(ctx, email, title, body)
}