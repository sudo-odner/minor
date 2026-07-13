package nats

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

// Описываем требования к сервису (чтобы не было циклической зависимости)
type UserSyncLogic interface {
	HandleRegistration(ctx context.Context, userID, email, username string) error
}

type UserConsumer struct {
	log     *zap.Logger
	js      jetstream.JetStream
	service UserSyncLogic
}

func NewUserConsumer(log *zap.Logger, js jetstream.JetStream, svc UserSyncLogic) *UserConsumer {
	return &UserConsumer{log: log, js: js, service: svc}
}

func (c *UserConsumer) Start(ctx context.Context) error {
	cons, err := c.js.CreateOrUpdateConsumer(ctx, "AUTH_STREAM", jetstream.ConsumerConfig{
		Durable:       "user-service-sync",
		FilterSubject: "auth.user.registered",
	})
	if err != nil {
		return err
	}

	cc, _ := cons.Consume(func(msg jetstream.Msg) {
		var event struct {
			UserID   string `json:"user_id"`
			Email    string `json:"email"`
			Username string `json:"username"`
		}
		json.Unmarshal(msg.Data(), &event)

		if err := c.service.HandleRegistration(ctx, event.UserID, event.Email, event.Username); err != nil {
			msg.Nak()
			return
		}
		msg.Ack()
	})

	<-ctx.Done()
	cc.Stop()
	return nil
}