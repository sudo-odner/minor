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
	client presencev1.PresenceServiceClient
	conn *grpc.ClientConn
}

func NewPresenceClient(addr string) (*Client, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("could not connect to presence service: %w", err)
	}

	return &Client{
		client: presencev1.NewPresenceServiceClient(conn),
		conn: conn,
	}, nil
}

func (c *Client) IsUserOnline(ctx context.Context, userID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2 * time.Second)
	defer cancel()

	resp, err := c.client.GetStatus(ctx, &presencev1.GetStatusRequest{
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