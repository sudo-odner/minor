package guild

import (
	"context"

	"github.com/google/uuid"
	communityv1 "github.com/sudo-odner/minor/backend/shared/pkg/pb/community/v1"
	"google.golang.org/grpc"
)

type Client struct {
	client communityv1.CommunityServiceClient
	conn   *grpc.ClientConn
}

func New(target string) (*Client, error) {
	const op = "client.grpc.guild.New"

	// TODO: enable when create comunication service
	// conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	return nil, fmt.Errorf("%s: %w", op, err)
	// }
	//
	// return &Client{
	// 	client: communityv1.NewCommunityServiceClient(conn),
	// 	conn:   conn,
	// }, nil
	return &Client{}, nil
}

// Закрыть gRPC соединение
func (c *Client) Close() error {
	return c.conn.Close()
}

// Получить побитовyю макску прав пользователя
func (c *Client) FetchPermission(ctx context.Context, userID, channelID uuid.UUID) (uint64, error) {
	// TODO: Implement logic
	return 0xFFFFFFFFFFFFFFFF, nil
}
