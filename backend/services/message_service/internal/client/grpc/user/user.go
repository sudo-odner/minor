package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	userv1 "github.com/sudo-odner/minor-shared/pkg/pb/user/v1"
	"google.golang.org/grpc"
)

type Client struct {
	client userv1.UserServiceClient
	conn   *grpc.ClientConn
}

func New(target string) (*Client, error) {
	const op = "client.grpc.user.New"

	// TODO: enable when create user service gRPC server
	// conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	return nil, fmt.Errorf("%s: %w", op, err)
	// }
	//
	// return &Client{
	// 	client: userv1.NewUserServiceClient(conn),
	// 	conn:   conn,
	// }, nil
	return &Client{}, nil
}

// Закрыть gRPC соединение
func (c *Client) Close() error {
	return c.conn.Close()
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
