package nuts

import (
	"time"

	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
)

// Subject
const (
	SubjectChannelCreated          = "community.channel.created"
	SubjectChannelUpdated          = "community.channel.updated"
	SubjectChannelDeleted          = "community.channel.deleted"
	SubjectChannelPositionsUpdated = "community.channel.positions_updated"

	SubjectUserDeleted = "user.deleted"
)

// Publish
type ChannelDTO struct {
	ID        string `json:"id"`
	ServerID  string `json:"server_id"`
	Name      string `json:"name"`
	Type      int    `json:"type"`
	ParentID  string `json:"parent_id"`
	Position  int    `json:"position"`
	CreatedAt string `json:"created_at"`
}

func toChannelDTO(ch *models.Channel) ChannelDTO {
	parentIDStr := ""
	if ch.ParentID != nil {
		parentIDStr = ch.ParentID.String()
	}

	return ChannelDTO{
		ID:        ch.ID.String(),
		ServerID:  ch.ServerID.String(),
		Name:      ch.Name,
		Type:      int(ch.Type),
		ParentID:  parentIDStr,
		Position:  ch.Position,
		CreatedAt: ch.CreatedAt.Format(time.RFC3339),
	}
}

type ShortChannelDTO struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
	Position int    `json:"position"`
}

func toShortChannelDTO(ch models.Channel) ShortChannelDTO {
	// TODO: Что-то нужно придумать не нравиться что переводим в ""
	parentIDStr := ""
	if ch.ParentID != nil {
		parentIDStr = ch.ParentID.String()
	}

	return ShortChannelDTO{
		ID:       ch.ID.String(),
		ParentID: parentIDStr,
		Position: ch.Position,
	}
}

type ChannelEvent struct {
	ServerID string     `json:"server_id"`
	Channel  ChannelDTO `json:"channel"`
}

type ChannelDeletedEvent struct {
	ServerID  string `json:"server_id"`
	ChannelID string `json:"channel_id"`
}

type ChannelPositionUpdate struct {
	ChannelID string `json:"channel_id"`
	Position  int    `json:"position"`
}

type ChannelPositionUpdateEvent struct {
	ServerID string            `json:"server_id"`
	Channels []ShortChannelDTO `json:"channels"`
}

// Subscriber
type UserDeleteEvent struct {
	UserID string `json:"user_id"`
}
