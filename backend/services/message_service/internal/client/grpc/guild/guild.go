package guild

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	guildv1 "github.com/sudo-odner/minor/backend/shared/pkg/pb/guild/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clinet struct {
	client guildv1.GuildServiceClient
	conn   *grpc.ClientConn
}

func New(target string) (*Clinet, error) {
	const op = "client.grpc.guild.New"

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Clinet{
		client: guildv1.NewGuildServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *Clinet) Close() error {
	return c.conn.Close()
}

func (c *Clinet) CanWrite(ctx context.Context, userID, channelID uuid.UUID) (bool, error) {
	// TODO: Implement
	return true, nil
}

func (c *Clinet) CanRead(ctx context.Context, userID, channelID uuid.UUID) (bool, error) {
	// TODO: Implement
	return true, nil
}
