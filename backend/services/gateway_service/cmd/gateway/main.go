package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/app"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/config"
	"go.uber.org/zap"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.App.Env)

	app := app.New(log, cfg)

	go func() {
		log.Info("Gateway Service started", zap.String("port", cfg.HTTP.Port))
		if err := app.GatewayHTTPServer.Run(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", zap.Error(err))
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Gateway Service...")

	shutdownCtx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	if err := app.GatewayHTTPServer.Stop(shutdownCtx); err != nil {
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
