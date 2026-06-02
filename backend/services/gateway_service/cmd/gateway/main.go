package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	authv1 "github.com/sudo-odner/minor-shared/pkg/pb/auth/v1"
	presencev1 "github.com/sudo-odner/minor-shared/pkg/pb/presence/v1"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/config"
	gatewayService "github.com/sudo-odner/minor/backend/services/gateway_service/internal/service/gateway"
	gatewayHandler "github.com/sudo-odner/minor/backend/services/gateway_service/internal/server/websocket"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.App.Env)

	// 2. Параметры конфигурации (обычно берутся из env/config.yaml)

	// 3. Подключение к gRPC сервисам
	authConn, err := grpc.Dial(cfg.GRPC.AuthService.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to connect auth service", zap.Error(err))
	}
	authClient := authv1.NewAuthServiceClient(authConn)

	presenceConn, err := grpc.Dial(cfg.GRPC.PresenceService.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to connect presence service", zap.Error(err))
	}
	presenceClient := presencev1.NewPresenceServiceClient(presenceConn)

	// 4. Подключение к NATS
	nc, err := nats.Connect(cfg.NATS.URL, nats.MaxReconnects(10), nats.ReconnectWait(time.Second*2))
	if err != nil {
		log.Fatal("failed to connect to NATS", zap.Error(err))
	}
	js, _ := jetstream.New(nc)

	// 5. Инициализация менеджера соединений (Hub)
	hub := gatewayService.NewHub(log)

	// 6. Подписки на NATS (Трансляция событий в сокеты)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// А. Трансляция ЧАТА (JetStream - надежная доставка)
	cons, _ := js.CreateOrUpdateConsumer(ctx, "CHAT_STREAM", jetstream.ConsumerConfig{
		Durable:       "gateway_durable_worker",
		FilterSubject: "chat.>",
	})
	go func() {
		iter, _ := cons.Consume(func(msg jetstream.Msg) {
			var rawData map[string]any
			if err := json.Unmarshal(msg.Data(), &rawData); err == nil {
				// Map created_at to create_at for frontend compatibility
				if createdAt, ok := rawData["created_at"]; ok {
					rawData["create_at"] = createdAt
				}
				
				wsPayload := map[string]any{
					"op": 1,
					"t":  "MESSAGE_CREATE",
					"d":  rawData,
				}
				
				wrappedBytes, err := json.Marshal(wsPayload)
				if err == nil {
					hub.Broadcast(wrappedBytes)
				}
			} else {
				hub.Broadcast(msg.Data())
			}
			msg.Ack()
		})
		<-ctx.Done()
		iter.Stop()
	}()

	// Б. Трансляция СООБЩЕСТВ (JetStream - надежная доставка событий каналов и участников)
	go func() {
		var commCons jetstream.Consumer
		var err error
		for {
			commCons, err = js.CreateOrUpdateConsumer(ctx, "COMMUNITY_STREAM", jetstream.ConsumerConfig{
				Durable:       "gateway_community_worker",
				FilterSubject: "community.>",
			})
			if err == nil {
				break
			}
			log.Warn("waiting for COMMUNITY_STREAM to be initialized...", zap.Error(err))
			time.Sleep(2 * time.Second)
		}

		iter, _ := commCons.Consume(func(msg jetstream.Msg) {
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
						hub.Broadcast(wrappedBytes)
					}
				}
			}
			msg.Ack()
		})
		<-ctx.Done()
		iter.Stop()
	}()

	// В. Трансляция СТАТУСОВ и ПЕЧАТИ (Core NATS - высокая скорость)
	nc.Subscribe("presence.status.updated", func(m *nats.Msg) {
		log.Info("received presence status update from NATS", zap.String("data", string(m.Data)))
		var rawData map[string]any
		if err := json.Unmarshal(m.Data, &rawData); err == nil {
			wsPayload := map[string]any{
				"op": 1,
				"t":  "PRESENCE_UPDATE",
				"d":  rawData,
			}
			if wrappedBytes, err := json.Marshal(wsPayload); err == nil {
				hub.Broadcast(wrappedBytes)
			}
		} else {
			hub.Broadcast(m.Data)
		}
	})
	nc.Subscribe("user.typing.*", func(m *nats.Msg) {
		var rawData map[string]any
		if err := json.Unmarshal(m.Data, &rawData); err == nil {
			wsPayload := map[string]any{
				"op": 1,
				"t":  "TYPING_START",
				"d":  rawData,
			}
			if wrappedBytes, err := json.Marshal(wsPayload); err == nil {
				hub.Broadcast(wrappedBytes)
			}
		} else {
			hub.Broadcast(m.Data)
		}
	})
	
	// 7. Настройка HTTP сервера и WebSocket эндпоинта
	handler := gatewayHandler.NewGatewayHandler(log, hub, authClient, presenceClient, nc)

	r := chi.NewRouter()
	r.Get("/gateway", handler.HandleWS) // Тот самый путь из вашего задания
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })

	server := &http.Server{
		Addr:    cfg.HTTP.Port,
		Handler: r,
	}

	// 8. ЗАПУСК
	go func() {
		log.Info("Gateway Service started", zap.String("port", cfg.HTTP.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", zap.Error(err))
		}
	}()

	// 9. Graceful Shutdown (ТЗ п. 5.2)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Gateway Service...")

	// Закрываем сокеты всех клиентов
	hub.CloseAll()

	// Останавливаем HTTP сервер
	shutdownCtx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("forced shutdown", zap.Error(err))
	}

	log.Info("Gateway Service stopped")
}

func setupLogger(env string) *zap.Logger {
	var log *zap.Logger
	var err error

	switch env {
	case envDev:
		log, err = zap.NewDevelopment()
		if err != nil {
			panic("failed to initialize development logger")
		}
	case envProd:
		log, err = zap.NewProduction()
		if err != nil {
			panic("failed to initialize production logger")
		}
	}

	return log
}
