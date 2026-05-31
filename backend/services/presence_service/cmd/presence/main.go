package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sudo-odner/minor/backend/services/presence_service/internal/app"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/lib/logger"
	"go.uber.org/zap"
)

func main() {
	// Init config and logger
	cfg := config.MustLoad()
	logg, err := logger.New(logger.Config{
		Env:         logger.Env(cfg.Env),
		ServiceName: "presence-service",
	})
	if err != nil {
		log.Fatalf("FATAL: failed init logger: %s", err)
	}

	// Init application
	logg.Info("starting init application")
	application, err := app.New(cfg, logg)
	if err != nil {
		logg.Fatal("failed init application", zap.Error(err))
	}

	// Run application in background
	go func() {
		application.Run()
	}()

	// Wait for terminate signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	logg.Info("stopping application")
	application.Stop(context.Background())
}
