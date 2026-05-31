package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/models"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/service/notifier"
	"go.uber.org/zap"
)

type NotificationConsumer struct {
	log *zap.Logger
	jetStream jetstream.JetStream
	notifier *notifier.Notifier
}

func NewNotificationConsumer(log *zap.Logger, jetStream jetstream.JetStream, notifier *notifier.Notifier) *NotificationConsumer{
	return &NotificationConsumer {
		log: log,
		jetStream: jetStream,
		notifier: notifier,
	}
}

func (c *NotificationConsumer) Start(ctx context.Context) error {
	log := c.log.With(
		zap.String("op:", "broker"),
	)
	
	cons, err := c.jetStream.CreateOrUpdateConsumer(ctx, "CHAT_STREAM", jetstream.ConsumerConfig{
		Durable: "notificaton-service-worker",
	})

	if err != nil {
		return fmt.Errorf("failed to start worker: %w", err)
	} 

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		var event models.ChatMessageCreated
		
		if err := json.Unmarshal(msg.Data(), &event); err != nil {
			log.Warn("failed to unmarshal data: %w", zap.Error(err))
			msg.Term()
			return
		}
		
		err := c.notifier.HandleChatMessage(ctx, event)
		if err != nil {
			log.Warn("gRPC error: %w", zap.Error(err))
			msg.Nak()
			return
		}

		msg.Ack()	
	})

	if err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	<-ctx.Done()
	cc.Stop()

	return nil
}