package gateway

import (
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

type Client struct {
	UserID   string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *Hub
	NatsConn *nats.Conn
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c.UserID)
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		// Если юзер прислал статус "печатаю", просто шлем в Core NATS
		// Это асинхронное взаимодействие из вашего списка
		c.NatsConn.Publish("user.typing", message)
	}
}

func (c *Client) WritePump() {
	for msg := range c.Send {
		c.Conn.WriteMessage(websocket.TextMessage, msg)
	}
}