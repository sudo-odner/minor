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
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/client/grpc/presence"
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
	log := setupLogger(cfg.Env)

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}

	presenceAddr := os.Getenv("PRESENCE_GRPC_ADDR")
	if presenceAddr == "" {
		presenceAddr = "presence_service:50051"
	}

	// 3. Подключение к NATS
	nc, err := nats.Connect(natsURL,
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

	// 4. Инициализация gRPC клиента к Presence Service
	pClient, err := presence.NewPresenceClient(presenceAddr)
	if err != nil {
		log.Fatal("failed to connect to presence service", zap.Error(err))
	}
	defer pClient.Close()

	uClient, _ := user.NewUserClient(userAddr)

	// 5. Сборка слоев (Dependency Injection)
	// Доставка (заглушка Firebase/Email)
	emailProvider := mail.NewSMTPProvider()

	// Бизнес-логика
	notifierSvc := notifyService.NewNotifier(pClient, emailProvider)

	// Воркер (Консьюмер NATS)
	notificationConsumer := broker.NewNotificationConsumer(log, js, notifierSvc)

	// 6. Контекст для Graceful Shutdown (п. 5.2 ТЗ)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 7. Запуск Воркера в отдельной горутине
	go func() {
		log.Info("starting notification consumer...")
		if err := notificationConsumer.Start(ctx); err != nil {
			log.Error("consumer stopped with error", zap.Error(err))
		}
	}()

	// 8. Запуск минимального HTTP сервера для Health Check
	// Это нужно для Docker/Kubernetes, чтобы знать, что контейнер жив
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	healthSrv := &http.Server{Addr: ":8081", Handler: mux}
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
