package nuts

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type CommunityConsumer struct {
	js     jetstream.JetStream
	log    *zap.Logger
	// Сюда можно добавить сервис для обновления данных в БД
}

func NewConsumer(js jetstream.JetStream, log *zap.Logger) *CommunityConsumer {
	return &CommunityConsumer{js: js, log: log}
}

func (c *CommunityConsumer) Start(ctx context.Context) error {
	// Например, слушаем удаление пользователя из Auth Service
	cons, _ := c.js.CreateOrUpdateConsumer(ctx, "AUTH_STREAM", jetstream.ConsumerConfig{
		Durable: "community-service-sync",
		FilterSubject: "auth.user.deleted",
	})

	cc, _ := cons.Consume(func(msg jetstream.Msg) {
		// Логика удаления пользователя со всех серверов
		msg.Ack()
	})
    
	<-ctx.Done()
	cc.Stop()
	return nil
}