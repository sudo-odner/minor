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
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/app"
	natsBroker "github.com/sudo-odner/minor/backend/services/auth_service/internal/broker/nats"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/repository/postgres"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/repository/redis"
	authHTTPHandler "github.com/sudo-odner/minor/backend/services/auth_service/internal/server/http/handler/auth"
	authGRPCHandler "github.com/sudo-odner/minor/backend/services/auth_service/internal/server/grpc/handler/auth"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/server/http/middleware/cors"
	authService "github.com/sudo-odner/minor/backend/services/auth_service/internal/service/auth"
	"go.uber.org/zap"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(envDev)

	nc, _ := nats.Connect(cfg.NATS.URL)
	js, _ := jetstream.New(nc)

	publisher := natsBroker.NewAuthPublisher(js)

	storagePath := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s", cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.DBName, os.Getenv("POSTGRES_PASSWORD"), cfg.Postgres.SSLMode)

	pgConn, err := postgres.New(context.Background(), storagePath)
	if err != nil {
		panic("failed to initialize Postgres connection")
	}

	rdb, err := redis.NewRedisClient(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, 0)
	if err != nil {
		panic("failed to initialize Redis connection")
	}

	redisConn := redis.NewSessionRepo(rdb)

	log.Info("starting authentication service")

	authService := authService.New(pgConn, redisConn, publisher, log, cfg.Auth)
	authHTTPHandler := authHTTPHandler.NewHTTPHandler(authService, log)
	authGRPCHandler := authGRPCHandler.NewGRPCHandler(authService, log)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(cors.NewCORS)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authHTTPHandler.Register(context.Background()))
		r.Post("/login", authHTTPHandler.Login(context.Background()))

		r.Post("/refresh", authHTTPHandler.RefreshToken(context.Background()))
		r.Post("/logout", authHTTPHandler.Logout(context.Background()))
		// r.Post("/logout-all", authHTTPHandler.LogoutAll(context.Background()))

		r.Post("/verify-internal", authHTTPHandler.VerifyInternal(context.Background()))
		// r.Post("/forgot-password")
		// r.Post("/reset-password")
	})

	application := app.New(log, cfg, router, authGRPCHandler)

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
