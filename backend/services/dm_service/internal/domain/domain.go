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
	UpdatedAt time.Time   `json:"updated_at"`
	CreatedAt time.Time   `json:"created_at"`
}

type UserMeta struct {
	ID   uuid.UUID
	Name string
}
