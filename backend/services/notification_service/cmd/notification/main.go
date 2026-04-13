package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/app"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/config"
	authHandler "github.com/sudo-odner/minor/backend/services/notification_service/internal/http-server/handler/auth"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/http-server/middleware/cors"
	"github.com/sudo-odner/minor/backend/services/notification_service/internal/repository/postgres"
	authService "github.com/sudo-odner/minor/backend/services/notification_service/internal/service/auth"
	"go.uber.org/zap"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(envDev)

	storagePath := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", cfg.PostgreConfig.Host, cfg.PostgreConfig.Port, cfg.PostgreConfig.Username, cfg.PostgreConfig.DBName, os.Getenv("POSTGRES_PASSWORD"), cfg.PostgreConfig.SSLMode)

	dbConn, err := postgres.New(context.Background(), storagePath)
	if err != nil {
		panic("failed to initialize DB connection")
	}

	log.Info("starting authentication service")

	authService := authService.New(dbConn, log)
	authHandler := authHandler.New(authService, log)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(cors.NewCORS)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/auth-service", func(r chi.Router) {
		r.Post("/register", authHandler.Register(context.Background()))
		r.Post("/login", authHandler.Login(context.Background()))
	})

	// router.Route("/token", func(r chi.Router) {
	// 	r.Post("/refresh")
	// })

	application := app.New(log, cfg, router)

	go func() {
		application.HTTPServer.Run()
	}()

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
