package app

import (
	"net/http"

	"github.com/go-chi/chi"
	httpapp "github.com/sudo-odner/minor/backend/services/gateway_service/internal/app/http"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/broker/nats"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/client/grpc/auth"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/client/grpc/presence"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/config"
	gatewayHandler "github.com/sudo-odner/minor/backend/services/gateway_service/internal/server/websocket"
	gatewayService "github.com/sudo-odner/minor/backend/services/gateway_service/internal/service/gateway"
	"go.uber.org/zap"
)

type App struct {
	GatewayHTTPServer *httpapp.GatewayHTTPServer
	log        *zap.Logger
}

func New(log *zap.Logger, cfg *config.Config) *App {
	const op = "app.New"

	log = log.With(
		zap.String("op", op),
	)

	log.Info("starting initialize app")

	hub := gatewayService.NewHub(log)
	
	nc, err := nats.New(log, cfg)
	if err != nil {
		log.Error("failed to initialize nats instance")
		panic(err)
	}

	nc.StartConsumers()

	authClient, err := auth.New(log, cfg)
	if err != nil {
		log.Error("failed to initialize auth grpc client", zap.Error(err))
		panic(err)
	}

	presenceClient, err := presence.New(log, cfg)
	if err != nil {
		log.Error("failed to initialize auth grpc client", zap.Error(err))
		panic(err)
	}
	
	handler := gatewayHandler.NewGatewayHandler(log, hub, authClient, presenceClient, nc)
	
	r := chi.NewRouter()
	r.Get("/gateway", handler.HandleWS)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })

	

	

	

	server := &http.Server{
		Addr:    cfg.HTTP.Port,
		Handler: r,
	}

	httpApp := httpapp.New(log, cfg)
	
	return &App{
		GatewayHTTPServer: httpApp,
		log:        log,
	}
}
