package handler

import (
	"net/http"
	"github.com/gorilla/websocket"
	authv1 "github.com/sudo-odner/minor-shared/pkg/pb/auth/v1"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/service/gateway"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // В ТЗ разрешить CORS
}

type GatewayHandler struct {
	hub        *gateway.Hub
	authClient authv1.AuthServiceClient
}

func (h *GatewayHandler) HandleGateway(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем токен из Query (ws://host/gateway?token=...)
	token := r.URL.Query().Get("token")

	// 2. gRPC вызов в Auth Service
	resp, err := h.authClient.VerifyToken(r.Context(), &authv1.VerifyTokenRequest{
		AccessToken: token,
	})
	if err != nil || !resp.IsValid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 3. Апгрейд до WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// 4. Создаем клиента и запускаем воркеры
	client := &service.Client{
		UserID: resp.UserId,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h.hub,
	}
	
	h.hub.Register(resp.UserId, client)

	go client.WritePump()
	go client.ReadPump()
}
