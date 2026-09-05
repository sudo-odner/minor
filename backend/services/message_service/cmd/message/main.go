package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sudo-odner/minor/backend/services/message_service/internal/app"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/config"
	"go.uber.org/zap"
)

func main() {
	// Setup config
	cfg := config.MustLoad()

	// Setup logger
	var logCfg zap.Config

	switch cfg.Env {
	case "local", "dev":
		logCfg = zap.NewDevelopmentConfig()
	default:
		logCfg = zap.NewProductionConfig()
	}

	logger, err := logCfg.Build()
	if err != nil {
		log.Fatalf("ERROR: falied iniitilzate logger")
	}

	// Starting application
	logger.Info("init application")
	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Error("falied initilizate application", zap.Error(err))
		os.Exit(1)
	}

	go func() {
		if err := application.Run(); err != nil {
			logger.Error("falied run application", zap.Error(err))
			return
		}
	}()

	// Gracefully shutting down
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	sig := <-stop

	logger.Info("shutting down application gracefully", zap.String("signal", sig.String()))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.Stop(ctx); err != nil {
		logger.Error("falied stop application", zap.Error(err))
	}
}
