package models

import (
	"time"

	"github.com/google/uuid"
)

type Server struct {
	ID        uuid.UUID
	Name      string
	OwnerID   uuid.UUID
	AvatarURL string
	CreatedAt time.Time
}

type Member struct {
	ServerID uuid.UUID
	UserID   uuid.UUID
	Nickname *string
	JoinedAt time.Time
}

type ChannelType int

const (
	ChannelTypeCategory ChannelType = 0
	ChannelTypeText     ChannelType = 1
	ChannelTypeVoice    ChannelType = 2
)

type OverrideType string

const (
	OverrideTypeRole OverrideType = "role"
	OverrideTypeUser OverrideType = "user"
)
