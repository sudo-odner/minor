package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/sudo-odner/min/backend/auth-service/internal/app"
	"github.com/sudo-odner/min/backend/auth-service/internal/config"
	"github.com/sudo-odner/min/backend/auth-service/internal/http-server/middleware/cors"
	"go.uber.org/zap"
)

const (
	envDev = "dev"
	envProd = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(envDev)

	log.Info("starting authentication service")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(cors.NewCORS)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	application := app.New(log, cfg, router)

	go func() {
		application.HTTPServer.Run()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	signal := <-stop

	log.Info("stopping application", zap.String("signal", signal.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
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