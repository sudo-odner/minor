package nats

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/service/gateway"
	"go.uber.org/zap"
)

type ConsumerManager struct {
	js  jetstream.JetStream
	log *zap.Logger
	hub *gateway.Hub
	// nodeID string // unique instance id
}

func NewConsumerManager(js jetstream.JetStream, log *zap.Logger, hub *gateway.Hub) *ConsumerManager {
	return &ConsumerManager{
		js:  js,
		log: log,
		hub: hub,
	}
}

func (cm *ConsumerManager) StartAllConsumers(ctx context.Context) error {
	const op = "broker.nats.StartAllConsumers"

	log := cm.log.With(
		zap.String("op", op),
	)

	// g, ctx := errgroup.WithContext(ctx)

	go func() {
		cons, err := cm.js.CreateOrUpdateConsumer(ctx, "CHAT_STREAM", jetstream.ConsumerConfig{
			Durable:       "gateway_durable_worker",
			FilterSubject: "chat.>",
		})

		if err != nil {
			log.Warn("failed to initialize chat stream", zap.Error(err))
		}

		iter, _ := cons.Consume(func(msg jetstream.Msg) {
			var rawData map[string]any
			if err := json.Unmarshal(msg.Data(), &rawData); err == nil {
				if createdAt, ok := rawData["created_at"]; ok {
					rawData["create_at"] = createdAt
				}

				log.Info("rawData", zap.Any("radData", rawData))

				wsPayload := map[string]any{
					"op": 1,
					"t":  "MESSAGE_CREATE",
					"d":  rawData,
				}

				wrappedBytes, err := json.Marshal(wsPayload)
				if err == nil {
					cm.hub.Broadcast(wrappedBytes)
				}
			} else {
				cm.hub.Broadcast(msg.Data())
			}
			msg.Ack()
		})
		<-ctx.Done()
		iter.Stop()
	}()

	go func() {
		cons, err := cm.js.CreateOrUpdateConsumer(ctx, "COMMUNITY_STREAM", jetstream.ConsumerConfig{
			Durable:       "gateway_community_worker",
			FilterSubject: "community.>",
		})
		if err == nil {
			log.Warn("failed to initialize community stream", zap.Error(err))
		}

		iter, err := cons.Consume(func(msg jetstream.Msg) {
			subject := msg.Subject()
			var rawData map[string]any
			if err := json.Unmarshal(msg.Data(), &rawData); err == nil {
				var tType string
				switch subject {
				case "community.channel.created":
					tType = "CHANNEL_CREATE"
				case "community.channel.updated":
					tType = "CHANNEL_UPDATE"
				case "community.channel.deleted":
					tType = "CHANNEL_DELETE"
				case "community.member.added":
					tType = "MEMBER_ADD"
				case "community.member.removed":
					tType = "MEMBER_REMOVE"
				}

				if tType != "" {
					wsPayload := map[string]any{
						"op": 1,
						"t":  tType,
						"d":  rawData,
					}
					wrappedBytes, err := json.Marshal(wsPayload)
					if err == nil {
						cm.hub.Broadcast(wrappedBytes)
					}
				}
			}
			msg.Ack()
		})
		<-ctx.Done()
		iter.Stop()
	}()

	go func() {
		cons, err := cm.js.CreateOrUpdateConsumer(ctx, "PRESENCE_STREAM")
		nc.Conn.Subscribe("presence.status.updated", func(m *nats.Msg) {
			log.Info("received presence status update from NATS", zap.String("data", string(m.Data)))
			var rawData map[string]any
			if err := json.Unmarshal(m.Data, &rawData); err == nil {
				wsPayload := map[string]any{
					"op": 1,
					"t":  "PRESENCE_UPDATE",
					"d":  rawData,
				}
				if wrappedBytes, err := json.Marshal(wsPayload); err == nil {
					cm.hub.Broadcast(wrappedBytes)
				}
			} else {
				cm.hub.Broadcast(m.Data)
			}
		})
	}()

	go func() {
		cons, err := cm.js.CreateOrUpdateConsumer(ctx, "")
		nc.Conn.Subscribe("user.typing.*", func(m *nats.Msg) {
			var rawData map[string]any
			if err := json.Unmarshal(m.Data, &rawData); err == nil {
				wsPayload := map[string]any{
					"op": 1,
					"t":  "TYPING_START",
					"d":  rawData,
				}
				if wrappedBytes, err := json.Marshal(wsPayload); err == nil {
					cm.hub.Broadcast(wrappedBytes)
				}
			} else {
				cm.hub.Broadcast(m.Data)
			}
		})
	}()

	return nil
}
