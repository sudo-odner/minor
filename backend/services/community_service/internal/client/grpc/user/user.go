package user

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Путь к сгенерированным файлам (замени 'minor' на свой модуль)
	userv1 "github.com/sudo-odner/minor-shared/pkg/pb/user/v1"
)

type Client struct {
	api  userv1.UserServiceClient
	conn *grpc.ClientConn
}

// NewUserClient создает подключение к User Service
func NewUserClient(addr string) (*Client, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	return &Client{
		api:  userv1.NewUserServiceClient(conn),
		conn: conn,
	}, nil
}

// GetBatchProfiles запрашивает сразу много профилей по списку ID.
// Это основной метод для обогащения списка участников сервера.
func (c *Client) GetBatchProfiles(ctx context.Context, userIDs []string) (map[string]*userv1.UserProfile, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.api.GetBatchProfiles(ctx, &userv1.GetBatchProfilesRequest{
		UserIds: userIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("grpc get batch profiles failed: %w", err)
	}

	return resp.Profiles, nil
}

// GetUserName запрашивает имя одного пользователя
func (c *Client) GetUserName(ctx context.Context, userID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := c.api.GetUserName(ctx, &userv1.GetUserNameRequest{
		UserId: userID,
	})
	if err != nil {
		return "", err
	}

	return resp.Username, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}