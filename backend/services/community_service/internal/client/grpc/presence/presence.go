package presence

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	presencev1 "github.com/sudo-odner/minor-shared/pkg/pb/presence/v1"
)

type Client struct {
	api  presencev1.PresenceServiceClient
	conn *grpc.ClientConn
}

func NewPresenceClient(addr string) (*Client, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to presence service: %w", err)
	}

	return &Client{
		api:  presencev1.NewPresenceServiceClient(conn),
		conn: conn,
	}, nil
}

func (c *Client) GetUserStatuses(ctx context.Context, userIDs []string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.api.GetUserStatuses(ctx, &presencev1.GetUserStatusesRequest{
		UserIds: userIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("grpc get user statuses failed: %w", err)
	}

	statuses := make(map[string]string)
	for id, p := range resp.Statuses {
		statuses[id] = p.Status.String()
	}

	return statuses, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
