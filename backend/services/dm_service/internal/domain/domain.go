package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChannelType string

const (
	ChannelTypeDM      = "DM"
	ChannelTypeDMGroup = "DM_GROUP"
)

type Channel struct {
	ID        uuid.UUID   `json:"id"`
	Type      ChannelType `json:"type"`
	Name      *string     `json:"name"`
	CreatedAt time.Time   `json:"created_at"`
}

type ChannelWithMembers struct {
	ID        uuid.UUID   `json:"id"`
	Type      ChannelType `json:"type"`
	Name      *string     `json:"name"`
	Members   []uuid.UUID `json:"users"`
	CreatedAt time.Time   `json:"created_at"`
}
