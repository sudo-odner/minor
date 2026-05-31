package presence

import (
	"context"
	"fmt"
	"time"

	presencev1 "github.com/sudo-odner/minor-shared/pkg/pb/presence/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api  presencev1.PresenceServiceClient // переименовал для удобства
	conn *grpc.ClientConn
}

func NewPresenceClient(addr string) (*Client, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// grpc.NewClient — это правильно (Go gRPC v1.60+)
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("could not connect to presence service: %w", err)
	}

	return &Client{
		api:  presencev1.NewPresenceServiceClient(conn),
		conn: conn,
	}, nil
}

// SetStatus — Обязательно нужен для Gateway Service
func (c *Client) SetStatus(ctx context.Context, userID string, status presencev1.UserStatus) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := c.api.SetStatus(ctx, &presencev1.SetStatusRequest{
		UserId: userID,
		Status: status,
	})
	if err != nil {
		return fmt.Errorf("failed to set status: %w", err)
	}
	return nil
}

// IsUserOnline — Оставил твою логику, она верна
func (c *Client) IsUserOnline(ctx context.Context, userID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := c.api.GetStatus(ctx, &presencev1.GetStatusRequest{
		UserId: userID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to get presence status: %w", err)
	}

	return resp.Presence.Status == presencev1.UserStatus_USER_STATUS_ONLINE, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}