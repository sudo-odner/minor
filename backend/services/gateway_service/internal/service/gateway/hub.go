package gateway

import (
	"sync"
	"go.uber.org/zap"

	"github.com/gorilla/websocket"
)

type Hub struct {
	// userID -> *Client
	clients map[string]*Client
	mu      sync.RWMutex
	log     *zap.Logger
}

func NewHub(log *zap.Logger) *Hub {
	return &Hub{
		clients: make(map[string]*Client),
		log:     log,
	}
}

func (h *Hub) Register(userID string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[userID] = client
}

func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, userID)
}

// Broadcast шлет сообщение всем, кто онлайн на данном сервере
func (h *Hub) Broadcast(data []byte) {
	h.mu.RLock()
	var toDelete []string
	for _, client := range h.clients {
		// Отправляем данные в канал WritePump этого клиента
		select {
		case client.Send <- data:
		default:
			toDelete = append(toDelete, client.UserID)
		}
	}
	h.mu.RUnlock()

	if len(toDelete) > 0 {
		h.mu.Lock()
		for _, userID := range toDelete {
			if client, ok := h.clients[userID]; ok {
				close(client.Send)
				delete(h.clients, userID)
			}
		}
		h.mu.Unlock()
	}
}

func (h *Hub) CloseAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.log.Info("closing all active websocket connections", zap.Int("count", len(h.clients)))

	for userID, client := range h.clients {
		// 1. Отправляем стандартный пакет закрытия WebSocket (Close Frame)
		// Это позволяет браузеру понять, что сервер уходит на покой
		err := client.Conn.WriteMessage(
			websocket.CloseMessage, 
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Server shutting down"),
		)
		if err != nil {
			h.log.Debug("failed to send close message", zap.String("user_id", userID), zap.Error(err))
		}

		// 2. Закрываем физическое TCP соединение
		client.Conn.Close()

		// 3. Удаляем из мапы
		delete(h.clients, userID)
	}
}