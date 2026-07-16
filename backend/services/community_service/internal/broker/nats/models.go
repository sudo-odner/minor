package nats

import (
	communityEvents "github.com/sudo-odner/minor-shared/pkg/nats/events/community"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
)

func toChannelDTO(ch *models.Channel) communityEvents.ChannelDTO {
	return communityEvents.ChannelDTO{
		ID:        ch.ID,
		ServerID:  ch.ServerID,
		Name:      ch.Name,
		Type:      int(ch.Type),
		ParentID:  ch.ParentID,
		Position:  ch.Position,
		CreatedAt: ch.CreatedAt,
	}
}

func toShortChannelDTO(ch models.Channel) communityEvents.ShortChannelDTO {
	return communityEvents.ShortChannelDTO{
		ID:       ch.ID,
		ParentID: ch.ParentID,
		Position: ch.Position,
	}
}
