package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sudo-odner/minor/backend/services/auth_service/internal/app"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/config"
	"go.uber.org/zap"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

func main() {
	var err error
	cfg := config.MustLoad()
	log := setupLogger(envDev)
	
	application := app.New(log, cfg)

	go func() {
		if err = application.HTTPServer.Run(); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err = application.GRPCServer.Run(); err != nil {
			panic(err)
		}
	}()

	// TODO: run a consumer and producers

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	signal := <-stop

	log.Info("stopping application", zap.String("signal", signal.String()))

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
