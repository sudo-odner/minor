package relationship

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	relationshipv1 "github.com/sudo-odner/minor-shared/pkg/pb/relationship/v1"
)

type Client struct {
	client relationshipv1.RelationshipServiceClient
}

func New(client relationshipv1.RelationshipServiceClient) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) FetchPermission(ctx context.Context, userID, channelID uuid.UUID) (authz.Permission, error) {
	resp, err := c.client.FetchPermission(ctx, &relationshipv1.FetchPermissionRequest{
		UserId:    userID.String(),
		ChannelId: channelID.String(),
	})
	if err != nil {
		return 0, fmt.Errorf("falied fetch user permission on channel: %w", err)
	}

	return authz.Permission(resp.GetPermissionMask()), nil
}
