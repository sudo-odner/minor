package notifier

import (
	"context"
	"fmt"

	messageEvents "github.com/sudo-odner/minor-shared/pkg/nats/events/message"
	authEvents "github.com/sudo-odner/minor-shared/pkg/nats/events/auth"
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

func (n *Notifier) HandleRegistration(ctx context.Context, event authEvents.UserRegisteredEvent) error {
	title := "Добро пожаловать в Minor!"
	body := fmt.Sprintf("Здравствуйте, %s!\n\nВы успешно зарегистрировались. Ваш ID: %s", 
		event.Email, event.UserID)

	return n.emailProvider.Send(ctx, event.Email, title, body)
}

// HandleLogin отправляет уведомление о входе
func (n *Notifier) HandleLogin(ctx context.Context, event authEvents.UserLoginSuccessEvent) error {
	// Сначала узнаем email и имя через gRPC, так как в событии Login их обычно нет (только ID)
	email, _ := n.userClient.GetUserEmail(ctx, event.UserID)
	username, _ := n.userClient.GetUserName(ctx, event.UserID)

	title := "Новый вход в аккаунт"
	body := fmt.Sprintf("Привет, %s!\n\nВ твой аккаунт вошли.\nIP: %s\nВремя: %s", 
		username, event.IP, event.Timestamp.Format("15:04:05 02.01.2006"))

	return n.emailProvider.Send(ctx, email, title, body)
}

func (n *Notifier) HandleChatMessage(ctx context.Context, event messageEvents.MessageCreatedEvent) error {
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

// func (n *Notifier) HandlePasswordReset(ctx context.Context, event events.PasswordResetRequestedEvent) error {
// 	title := "Восстановление пароля в Minor"
	
// 	// Используем твой стиль из примера SMTP клиента
// 	body := fmt.Sprintf(
// 		"Здравствуйте, %s!\n\n"+
// 			"Был получен запрос на сброс пароля для вашего аккаунта.\n"+
// 			"Ваш проверочный код: %s\n\n"+
// 			"Этот код действителен в течение 15 минут. Если вы не запрашивали сброс, просто проигнорируйте это письмо.",
// 		event.Username, event.Code,
// 	)

// 	// Отправляем через уже реализованный Gmail/SMTP провайдер
// 	err := n.emailProvider.Send(ctx, event.Email, title, body)
// 	if err != nil {
// 		return fmt.Errorf("failed to send reset email: %w", err)
// 	}

// 	return nil
// }

func (n *Notifier) HandlePasswordReset(ctx context.Context, event events.PasswordResetRequestedEvent) error {
    // Формируем ссылку на фронтенд (localhost для разработки)
    resetLink := fmt.Sprintf("http://localhost/reset-password?token=%s", event.Code)

    title := "Восстановление пароля"
    body := fmt.Sprintf(
        "Здравствуйте, %s!\n\nДля сброса пароля перейдите по ссылке:\n%s\n\nСсылка действует 15 минут.",
        event.Username, resetLink,
    )

    return n.emailProvider.Send(ctx, event.Email, title, body)
}