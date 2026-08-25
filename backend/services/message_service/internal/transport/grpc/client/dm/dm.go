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

func (c *Client) CheckChannelExists(ctx context.Context, channelID uuid.UUID) (bool, error) {
	resp, err := c.client.CheckChannelExists(ctx, &dmv1.CheckChannelExistsRequest{
		ChannelId: channelID.String(),
	})
	if err != nil {
		return false, fmt.Errorf("falied check channel exist in community service: %w", err)
	}

	return resp.GetExists(), nil
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
