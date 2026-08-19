package community

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	communityv1 "github.com/sudo-odner/minor-shared/pkg/pb/community/v1"
)

type CommunityClient struct {
	client communityv1.CommunityServiceClient
}

func New(client communityv1.CommunityServiceClient) *CommunityClient {
	return &CommunityClient{
		client: client,
	}
}

func (c *CommunityClient) FetchPermission(ctx context.Context, userID, channelID uuid.UUID) (authz.Permission, error) {
	resp, err := c.client.FetchPermission(ctx, &communityv1.FetchPermissionRequest{
		UserId:    userID.String(),
		ChannelId: channelID.String(),
	})
	if err != nil {
		return 0, fmt.Errorf("falied fetch user permission on channel: %w", err)
	}

	return authz.Permission(resp.GetPermissionMask()), nil
}
