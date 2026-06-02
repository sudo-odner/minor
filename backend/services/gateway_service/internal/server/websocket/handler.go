package ws

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	authv1 "github.com/sudo-odner/minor-shared/pkg/pb/auth/v1"
	presencev1 "github.com/sudo-odner/minor-shared/pkg/pb/presence/v1"
	service "github.com/sudo-odner/minor/backend/services/gateway_service/internal/service/gateway"
)

type GatewayHandler struct {
	log            *zap.Logger
	hub            *service.Hub
	authClient     authv1.AuthServiceClient
	presenceClient presencev1.PresenceServiceClient
	natsConn       *nats.Conn
	upgrader       websocket.Upgrader
}

func NewGatewayHandler(
	log *zap.Logger,
	hub *service.Hub,
	auth authv1.AuthServiceClient,
	presence presencev1.PresenceServiceClient,
	nats *nats.Conn,
) *GatewayHandler {
	return &GatewayHandler{
		log:            log,
		hub:            hub,
		authClient:     auth,
		presenceClient: presence,
		natsConn:       nats,
		upgrader: websocket.Upgrader{
    		ReadBufferSize:  1024,
      		WriteBufferSize: 1024,
		    CheckOrigin: func(r *http.Request) bool {
		        return true
		    },
		},
	}
}

func (h *GatewayHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	const path = "server.websocket.HandleWS"
	
	log := h.log.With(
		zap.String("path", path),
		zap.String("req-id", middleware.GetReqID(r.Context())),
	)

	log.Info("starting handle websocket")
	
	// 1. Аутентификация (берем токен из Query-параметра ?token=...)
	token := r.URL.Query().Get("token")
	if token == "" {
		h.log.Warn("connection attempt without token")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Info("got token", zap.String("token", token))

	// Вызов gRPC Auth Service
	authResp, err := h.authClient.VerifyToken(r.Context(), &authv1.VerifyTokenRequest{
		AccessToken: token,
	})
	if err != nil || !authResp.IsValid {
		h.log.Warn("invalid token", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Info("success auth grpc call")

	// 2. Апгрейд протокола до WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("failed to upgrade to websocket", zap.Error(err))
		return
	}

	log.Info("conn upgraded")	

	// 3. Регистрация статуса ONLINE в Presence Service (через gRPC)
	_, err = h.presenceClient.SetStatus(r.Context(), &presencev1.SetStatusRequest{
		UserId: authResp.UserId,
		Status: presencev1.UserStatus_USER_STATUS_ONLINE,
	})
	if err != nil {
		h.log.Error("failed to set online status", zap.Error(err))
		// Продолжаем работу, даже если Presence упал, сокет важнее
	}

	log.Info("presence status update")

	// 4. Создание объекта Client
	// Передаем все необходимые зависимости, включая presenceClient для дефера в ReadPump
	client := service.NewClient(
		authResp.UserId,
		conn,
		h.hub,
		h.natsConn,
		h.presenceClient, // <-- важно для OFFLINE статуса при дисконнекте
	)

	// 5. Регистрируем клиента в хабе
	h.hub.Register(authResp.UserId, client)

	// 6. Запускаем жизненный цикл сокета
	go client.WritePump()
	go client.ReadPump()

	h.log.Info("new websocket connection", zap.String("user_id", authResp.UserId))
}
