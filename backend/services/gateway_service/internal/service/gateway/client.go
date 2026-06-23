package gateway

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	presencev1 "github.com/sudo-odner/minor-shared/pkg/pb/presence/v1"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/models"
)

type Client struct {
	UserID         string
	Conn           *websocket.Conn
	Send           chan []byte
	hub            *Hub
	natsConn       *nats.Conn
	presenceClient presencev1.PresenceServiceClient
}

type TypingData struct {
	ChannelID string `json:"channel_id"`
}

func NewClient(userID string, conn *websocket.Conn,	hub *Hub, nats *nats.Conn, presence presencev1.PresenceServiceClient) *Client {
	return &Client{
		UserID:         userID,
		Conn:           conn,
		Send:           make(chan []byte, 256),
		hub:            hub,
		natsConn:       nats,
		presenceClient: presence,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c.UserID)

		_, err := c.presenceClient.SetStatus(context.Background(), &presencev1.SetStatusRequest{
			UserId: c.UserID,
			Status: presencev1.UserStatus_USER_STATUS_OFFLINE, // Используем константу из proto
		})
		if err != nil {
			fmt.Printf("failed to set offline status: %v\n", err)
		}

		c.Conn.Close()
	}()

	for {
		_, payload, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var wsMsg models.WSPayload
		json.Unmarshal(payload, &wsMsg)

		if wsMsg.Op == 3 {
			var data TypingData
			if err := json.Unmarshal(wsMsg.D, &data); err != nil {
				continue
			}

			subject := fmt.Sprintf("user.typing.%s", data.ChannelID)
			c.natsConn.Publish(subject, payload)
		}
	}
}

func (c *Client) WritePump() {
	for msg := range c.Send {
		c.Conn.WriteMessage(websocket.TextMessage, msg)
	}
}
