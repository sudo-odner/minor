package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/broker"
	presenceClient "github.com/sudo-odner/minor/backend/services/notification_service/internal/client/grpc/presence"
	userClient "github.com/sudo-odner/minor/backend/services/notification_service/internal/client/grpc/user"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/lib/mail"
	notifyService "github.com/sudo-odner/minor/backend/services/notification_service/internal/service/notifier"
	"go.uber.org/zap"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.App.Env)

	// 3. Подключение к NATS
	nc, err := nats.Connect(cfg.Nats.URL,
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		log.Fatal("failed to connect to NATS", zap.Error(err))
	}
	defer nc.Close()

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal("failed to init JetStream", zap.Error(err))
	}

	initStreams(context.Background(), js, log)

	// 4. Инициализация gRPC клиента к Presence Service
	presenceClient, err := presenceClient.NewPresenceClient(cfg.GRPC.PresenceService.Address)
	if err != nil {
		log.Fatal("failed to connect to presence service", zap.Error(err))
	}
	defer presenceClient.Close()

	userClient, _ := userClient.NewUserClient(cfg.GRPC.PresenceService.Address)

	// 5. Сборка слоев (Dependency Injection)
	// Доставка (заглушка Firebase/Email)
	emailProvider := mail.NewSMTPProvider()

	// Бизнес-логика
	notifierSvc := notifyService.NewNotifier(presenceClient, userClient, emailProvider)

	// Воркер (Консьюмер NATS)
	notificationConsumer := broker.NewNotificationConsumer(log, js, notifierSvc)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 7. Запуск Воркера в отдельной горутине
	go func() {
		if err := notificationConsumer.StartChatConsumer(ctx); err != nil {
			log.Error("chat consumer error", zap.Error(err))
		}
	}()

	// Поток 2: Auth (регистрация и логин)
	go func() {
		if err := notificationConsumer.StartAuthConsumer(ctx); err != nil {
			log.Error("auth consumer error", zap.Error(err))
		}
	}()

	// 8. Запуск минимального HTTP сервера для Health Check
	// Это нужно для Docker/Kubernetes, чтобы знать, что контейнер жив
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	healthSrv := &http.Server{Addr: cfg.HTTP.Port, Handler: mux}
	go func() {
		if err := healthSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("health server error", zap.Error(err))
		}
	}()

	// 9. Ожидание сигнала завершения
	<-ctx.Done()
	log.Info("shutting down notification service...")

	// Даем воркеру и серверу 5 секунд на мягкое завершение
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := healthSrv.Shutdown(shutdownCtx); err != nil {
		log.Error("health server forced shutdown", zap.Error(err))
	}

	log.Info("notification service stopped")
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

func initStreams(ctx context.Context, js jetstream.JetStream, logger *zap.Logger) {
	// Список потоков, которые нам нужны
	streams := []jetstream.StreamConfig{
		{
			Name:     "CHAT_STREAM",
			Subjects: []string{"chat.message.>"},
		},
		{
			Name:     "AUTH_STREAM",
			// Subjects: []string{"auth.user.>", "auth.password.>"},
			Subjects: []string{"auth.>"},
		},
		{
			Name:     "COMMUNITY_STREAM",
			Subjects: []string{"community.>"},
		},
		{
			Name:     "USER_STREAM",
			Subjects: []string{"user.>"},
		},
		{
			Name:     "RELATIONSHIP_STREAM",
			Subjects: []string{"relationship.>"},
		},
	}

	for _, cfg := range streams {
		_, err := js.CreateOrUpdateStream(ctx, cfg)
		if err != nil {
			logger.Fatal("failed to create stream", 
				zap.String("stream", cfg.Name), 
				zap.Error(err),
			)
		}
		logger.Info("stream initialized", zap.String("stream", cfg.Name))
	}
}