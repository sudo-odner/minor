package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sudo-odner/minor/backend/services/user_service/internal/app"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/lib/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.MustLoad()

	logger, err := logger.New(logger.Config{
		Env:         logger.EnvLocal,
		ServiceName: "user_service",
	})
	if err != nil {
		log.Fatal("faild initialize logger")
	}

	logger.Info("starting user service")

	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatal("faild initalize application", zap.Error(err))
	}

	go application.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	var signal os.Signal
	select {
	case err := <-application.ErrChan:
		logger.Fatal("faild user server", zap.Error(err))
	case signal = <-stop:
		logger.Info("shutting down")
	}

	logger.Info("stopping application", zap.String("signal", signal.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := application.Stop(ctx); err != nil {
		logger.Fatal("failed stoped", zap.Error(err))
	}

	logger.Info("stopped application")
}
