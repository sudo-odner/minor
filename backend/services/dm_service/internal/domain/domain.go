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
	ID        uuid.UUID
	Type      ChannelType
	Name      *string
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}
