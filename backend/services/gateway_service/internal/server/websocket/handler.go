package ws

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	presencev1 "github.com/sudo-odner/minor-shared/pkg/pb/presence/v1"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/client/grpc/auth"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/client/grpc/presence"
	service "github.com/sudo-odner/minor/backend/services/gateway_service/internal/service/gateway"
)

type GatewayHandler struct {
	log            *zap.Logger
	hub            *service.Hub
	authClient     *auth.AuthGRPCClient
	presenceClient *presence.PresenceGRPCClient
	natsConn       *nats.Conn
	upgrader       websocket.Upgrader
}

func NewGatewayHandler(log *zap.Logger, hub *service.Hub, auth *auth.AuthGRPCClient, presence *presence.PresenceGRPCClient, nats *nats.Conn) *GatewayHandler {
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

	token := r.URL.Query().Get("token")
	if token == "" {
		h.log.Warn("connection attempt without token")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Info("got token", zap.String("token", token))

	authResp, err := h.authClient.VerifyToken(r.Context(), token)

	if err != nil || !authResp.IsValid {
		h.log.Warn("invalid token", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Info("success auth grpc call")

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("failed to upgrade to websocket", zap.Error(err))
		return
	}

	log.Info("conn upgraded")

	client := service.NewClient(
		authResp.UserId,
		conn,
		h.hub,
		h.natsConn,
		h.presenceClient,
	)

	h.hub.Register(authResp.UserId, client)

	go client.WritePump()
	go client.ReadPump()

	err = h.presenceClient.SetStatus(r.Context(), authResp.UserId, presencev1.UserStatus_USER_STATUS_ONLINE)
	if err != nil {
		h.log.Error("failed to set online status", zap.Error(err))
	}

	h.log.Info("new websocket connection", zap.String("user_id", authResp.UserId))
}
