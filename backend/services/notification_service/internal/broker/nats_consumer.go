package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/sudo-odner/minor-shared/pkg/events"
	service "github.com/sudo-odner/minor/backend/services/notification_service/internal/service/notifier"
	"go.uber.org/zap"
)

type NotificationConsumer struct {
	log       *zap.Logger
	jetStream jetstream.JetStream
	notifier  *service.Notifier
}

func NewNotificationConsumer(log *zap.Logger, js jetstream.JetStream, n *service.Notifier) *NotificationConsumer {
	return &NotificationConsumer{
		log:       log,
		jetStream: js,
		notifier:  n,
	}
}

func (c *NotificationConsumer) StartChatConsumer(ctx context.Context) error {
	cons, err := c.jetStream.CreateOrUpdateConsumer(ctx, "CHAT_STREAM", jetstream.ConsumerConfig{
		Durable:       "notification-chat-worker",
		FilterSubject: "chat.message.>",
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to create chat consumer: %w", err)
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		var event events.MessageCreatedEvent
		if err := json.Unmarshal(msg.Data(), &event); err != nil {
			c.log.Error("failed to unmarshal chat event", zap.Error(err))
			msg.Term()
			return
		}

		if err := c.notifier.HandleChatMessage(ctx, event); err != nil {
			c.log.Error("failed to process chat notification", zap.Error(err))
			msg.Nak()
			return
		}
		msg.Ack()
	})

	if err != nil {
		return err
	}

	<-ctx.Done()
	cc.Stop()
	return nil
}

// StartAuthConsumer — Метод для обработки регистрации и входа
func (c *NotificationConsumer) StartAuthConsumer(ctx context.Context) error {
	cons, err := c.jetStream.CreateOrUpdateConsumer(ctx, "AUTH_STREAM", jetstream.ConsumerConfig{
		Durable:       "notification-auth-worker",
		FilterSubject: "auth.user.>",
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to create auth consumer: %w", err)
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		subject := msg.Subject()
		var err error

		switch subject {
		case "auth.user.registered":
			var event events.UserRegisteredEvent
			if err = json.Unmarshal(msg.Data(), &event); err == nil {
				err = c.notifier.HandleRegistration(ctx, event)
			}
		case "auth.user.login_success":
			var event events.UserLoginSuccessEvent
			if err = json.Unmarshal(msg.Data(), &event); err == nil {
				err = c.notifier.HandleLogin(ctx, event)
			}
		case "auth.password.reset_requested":
			var event events.PasswordResetRequestedEvent
			if err = json.Unmarshal(msg.Data(), &event); err == nil {
				err = c.notifier.HandlePasswordReset(ctx, event)
			}
		}

		if err != nil {
			msg.Nak()
			return
		}
		msg.Ack()
	})

	if err != nil {
		return err
	}

	<-ctx.Done()
	cc.Stop()
	return nil
}
