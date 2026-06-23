package presence

import (
	"context"
	"fmt"
	"time"

	presencev1 "github.com/sudo-odner/minor-shared/pkg/pb/presence/v1"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PresenceGRPCClient struct {
	api  presencev1.PresenceServiceClient
	conn *grpc.ClientConn
}

func New(log *zap.Logger, cfg *config.Config) (*PresenceGRPCClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	presenceConn, err := grpc.NewClient(cfg.GRPC.PresenceService.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("could not connect to presence service: %w", err)
	}

	return &PresenceGRPCClient{
		api:  presencev1.NewPresenceServiceClient(presenceConn),
		conn: presenceConn,
	}, nil
}

func (c *PresenceGRPCClient) SetStatus(ctx context.Context, userID string, status presencev1.UserStatus) error {
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

func (c *PresenceGRPCClient) IsUserOnline(ctx context.Context, userID string) (bool, error) {
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

func (c *PresenceGRPCClient) Close() error {
	return c.conn.Close()
}