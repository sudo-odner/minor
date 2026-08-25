package dm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	dmv1 "github.com/sudo-odner/minor-shared/pkg/pb/dm/v1"
)

type Client struct {
	client dmv1.DMServiceClient
}

func New(client dmv1.DMServiceClient) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) FetchPermission(ctx context.Context, userID, channelID uuid.UUID) (authz.Permission, error) {
	resp, err := c.client.FetchPermission(ctx, &dmv1.FetchPermissionRequest{
		UserId:    userID.String(),
		ChannelId: channelID.String(),
	})
	if err != nil {
		return 0, fmt.Errorf("falied fetch user permission on channel: %w", err)
	}

	return authz.Permission(resp.GetPermissionMask()), nil
}

// TODO: CheckChannelExists(ctx context.Context, channelID uuid.UUID) (bool, error)
