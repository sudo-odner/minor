package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sudo-odner/minor/backend/service/dm_service/internal/app"
	"github.com/sudo-odner/minor/backend/service/dm_service/internal/config"
)

func main() {
	// Config
	cfg := config.MustLoad()

	// Logger
	var logHandler slog.Handler
	switch cfg.Env {
	case "local":
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case "dev":
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	default:
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	logger := slog.New(logHandler)

	// Applicatioin
	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Error("falied initilizate application", slog.String("error", err.Error()))
		os.Exit(1)
	}

	go func() {
		if err := application.Run(); err != nil {
			logger.Error("falied run applicaiton", slog.String("error", err.Error()))
		}
	}()

	// GC
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	sig := <-stop
	logger.Info("shutting down application gracefully", slog.String("signal", sig.String()))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.Stop(ctx); err != nil {
		logger.Error("falied stop application gracefully", slog.String("error", err.Error()))
		return
	}
	logger.Info("application gracefully stopped")
}
