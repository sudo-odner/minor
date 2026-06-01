package user

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userv1 "github.com/sudo-odner/minor-shared/pkg/pb/user/v1"
)

type Client struct {
	api  userv1.UserServiceClient
	conn *grpc.ClientConn
}

// NewUserClient создает новое gRPC подключение к User Service
func NewUserClient(addr string) (*Client, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Используем grpc.NewClient (стандарт для новых версий Go gRPC)
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("could not connect to user service: %w", err)
	}

	return &Client{
		api:  userv1.NewUserServiceClient(conn),
		conn: conn,
	}, nil
}

// GetUserEmail запрашивает email пользователя по его ID
func (c *Client) GetUserEmail(ctx context.Context, userID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := c.api.GetUserEmail(ctx, &userv1.GetUserEmailRequest{
		UserId: userID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get user email: %w", err)
	}

	return resp.Email, nil
}

// GetUserName запрашивает никнейм пользователя по его ID
func (c *Client) GetUserName(ctx context.Context, userID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := c.api.GetUserName(ctx, &userv1.GetUserNameRequest{
		UserId: userID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get username: %w", err)
	}

	return resp.Username, nil
}

// Close закрывает соединение при выключении сервиса
func (c *Client) Close() error {
	return c.conn.Close()
}