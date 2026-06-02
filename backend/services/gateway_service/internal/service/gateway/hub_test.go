package gateway_test

import (
	"testing"
	"time"

	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/service/gateway"
	"go.uber.org/zap"
)

func TestHub_RegisterAndUnregister(t *testing.T) {
	logger := zap.NewNop()
	hub := gateway.NewHub(logger)

	userID := "user-123"
	client := gateway.NewClient(userID, nil, hub, nil, nil)

	// Test Register
	hub.Register(userID, client)

	// Test Unregister
	hub.Unregister(userID)
}

func TestHub_Broadcast(t *testing.T) {
	logger := zap.NewNop()
	hub := gateway.NewHub(logger)

	userID1 := "user-1"
	client1 := gateway.NewClient(userID1, nil, hub, nil, nil)

	userID2 := "user-2"
	client2 := gateway.NewClient(userID2, nil, hub, nil, nil)

	hub.Register(userID1, client1)
	hub.Register(userID2, client2)

	message := []byte("hello world")

	// Run Broadcast
	hub.Broadcast(message)

	// Verify client1 received the broadcast
	select {
	case msg := <-client1.Send:
		if string(msg) != string(message) {
			t.Errorf("client1 expected message %s, got %s", message, msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for client1 to receive message")
	}

	// Verify client2 received the broadcast
	select {
	case msg := <-client2.Send:
		if string(msg) != string(message) {
			t.Errorf("client2 expected message %s, got %s", message, msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for client2 to receive message")
	}
}

func TestHub_BroadcastBlockedChannel(t *testing.T) {
	logger := zap.NewNop()
	hub := gateway.NewHub(logger)

	userID := "user-blocked"
	client := gateway.NewClient(userID, nil, hub, nil, nil)
	// Force the channel to be full by filling it (Send has size 256, let's fill it)
	for i := 0; i < 256; i++ {
		client.Send <- []byte("fill")
	}

	hub.Register(userID, client)

	// Broadcast should trigger default case for blocked channel, close it, and remove client from hub
	hub.Broadcast([]byte("blocked_trigger"))

	// Verify client's channel is closed
	select {
	case _, ok := <-client.Send:
		if ok {
			// Read the filled messages first
			for len(client.Send) > 0 {
				<-client.Send
			}
			// Now it should close
			_, ok2 := <-client.Send
			if ok2 {
				t.Error("expected channel to be closed")
			}
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for channel close check")
	}
}
