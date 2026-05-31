package nuts

import (
	"github.com/sudo-odner/minor-shared/pkg/events"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
)

func toChannelDTO(ch *models.Channel) events.ChannelDTO {
	return events.ChannelDTO{
		ID:        ch.ID,
		ServerID:  ch.ServerID,
		Name:      ch.Name,
		Type:      int(ch.Type),
		ParentID:  ch.ParentID,
		Position:  ch.Position,
		CreatedAt: ch.CreatedAt,
	}
}

func toShortChannelDTO(ch models.Channel) events.ShortChannelDTO {
	return events.ShortChannelDTO{
		ID:       ch.ID,
		ParentID: ch.ParentID,
		Position: ch.Position,
	}
}
