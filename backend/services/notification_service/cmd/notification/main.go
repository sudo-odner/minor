package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/app"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/broker"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/client/grpc/presence"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/delivery"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/service/notifier"
	"go.uber.org/zap"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(envDev)

	log.Info("starting authentication service")
	
	natsConn, err := nats.Connect("nats://nats:4222")
	if err != nil {
		log.Fatal("failed to initialize nats client:", zap.Error(err))
	}

	jetStreamInstance, err := jetstream.New(natsConn)
	if err != nil {
		log.Fatal("failed to initialize jetstream instance:", zap.Error(err))
	}

	presenceClient, err := presence.NewPresenceClient(cfg.ClientDomain)
	if err != nil {
		log.Fatal("failed to initialize presence grpc client:", zap.Error(err))
	}

	notifier := notifier.NewNotifier(presenceClient, &delivery.FCMMock{})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	consumer := broker.NewNotificationConsumer(log, jetStreamInstance, notifier)
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Fatal("failed to start consumer:", zap.Error(err))
		}
	}()

	application := app.New(log, cfg)

	go func() {
		application.HTTPServer.Run()
	}()

	<-ctx.Done()

	log.Info("stopping application", zap.String("signal", ctx.Err().Error()))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	application.HTTPServer.Stop(ctx)

	log.Info("application stopped")
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
