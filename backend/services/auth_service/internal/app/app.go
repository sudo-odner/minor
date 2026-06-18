package app

import (
	"context"
	"fmt"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	grpcapp "github.com/sudo-odner/minor/backend/services/auth_service/internal/app/grpc"
	httpapp "github.com/sudo-odner/minor/backend/services/auth_service/internal/app/http"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/config"
	"go.uber.org/zap"

	natsapp "github.com/sudo-odner/minor/backend/services/auth_service/internal/broker/nats"
	natsProducer "github.com/sudo-odner/minor/backend/services/auth_service/internal/broker/nats/producer"
	// natsConsumer "github.com/sudo-odner/minor/backend/services/auth_service/internal/broker/nats/consumer"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/repository/postgres"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/repository/redis"
	authGRPCHandler "github.com/sudo-odner/minor/backend/services/auth_service/internal/server/grpc/handler/auth"
	authHTTPHandler "github.com/sudo-odner/minor/backend/services/auth_service/internal/server/http/handler/auth"
	authService "github.com/sudo-odner/minor/backend/services/auth_service/internal/service/auth"
	passwordHTTPHandler"github.com/sudo-odner/minor/backend/services/auth_service/internal/server/http/handler/password"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/server/http/middleware/cors"
)

type App struct {
	HTTPServer *httpapp.App
	GRPCServer *grpcapp.App
	log        *zap.Logger
}

func New(log *zap.Logger, cfg *config.Config) *App {
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

	natsApp, err := natsapp.New(log, cfg)
	if err != nil {
		panic("failed to initialize NATS")
	}

	log.Info("starting authentication service")

	publisher := natsProducer.NewAuthPublisher(natsApp.JS)
	
	authService := authService.New(pgConn, redisConn, redisConn, publisher, log, cfg.Auth)
	authHTTPHandler := authHTTPHandler.NewHTTPHandler(authService, log)
	authGRPCHandler := authGRPCHandler.NewGRPCHandler(authService, log)

	passwordHTTPHandler := passwordHTTPHandler.NewHTTPHandler(log, authService)

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

		r.HandleFunc("/verify-internal", authHTTPHandler.VerifyInternal(context.Background()))
		r.Post("/forgot-password", passwordHTTPHandler.ForgotPassword(context.Background()))
		r.Post("/reset-password", passwordHTTPHandler.ResetPassword(context.Background()))
	})

	httpApp := httpapp.New(log, cfg, router)
	grpcApp := grpcapp.New(log, authGRPCHandler, cfg.GRPC.Port)

	return &App{
		HTTPServer: httpApp,
		GRPCServer: grpcApp,
		log:        log,
	}
}
