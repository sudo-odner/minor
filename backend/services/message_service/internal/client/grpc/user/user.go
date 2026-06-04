package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	userv1 "github.com/sudo-odner/minor-shared/pkg/pb/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client userv1.UserServiceClient
	conn   *grpc.ClientConn
}

func New(target string) (*Client, error) {
	const op = "client.grpc.user.New"

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	
	return &Client{
		client: userv1.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

// Закрыть gRPC соединение
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Получить побитовyю макску прав пользователя
func (c *Client) FetchPermission(ctx context.Context, userID, channelID uuid.UUID) (authz.Permission, error) {
	// TODO: Implement logic
	return 0xFFFFFFFFFFFFFFFF, nil
}

// Пренадлежит ли канал сервису
func (c *Client) CheckChannelExists(ctx context.Context, channelID uuid.UUID) (bool, error) {
	// TODO: Immplemet logic
	return true, nil
}

func (c *Client) GetBatchProfiles(ctx context.Context, userIDs []string) (map[string]*userv1.UserProfile, error) {
	resp, err := c.client.GetBatchProfiles(ctx, &userv1.GetBatchProfilesRequest{
		UserIds: userIDs,
	})
	if err != nil {
		return nil, err
	}

	return resp.GetProfiles(), nil
}

func (c *Client) GetUserName(ctx context.Context, userID string) (string, error) {
	// ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	// defer cancel()

	resp, err := c.client.GetUserName(ctx, &userv1.GetUserNameRequest{
		UserId: userID,
	})
	if err != nil {
		return "", fmt.Errorf("grpc get username failed: %w", err)
	}

	return resp.GetUsername(), nil
}