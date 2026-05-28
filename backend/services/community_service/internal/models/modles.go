package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/shared/pkg/authz"
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

type Channel struct {
	ID        uuid.UUID
	ServerID  uuid.UUID
	Name      string
	Type      ChannelType
	ParentID  *uuid.UUID
	Position  int
	CreatedAt time.Time
}

type OverrideType string

const (
	OverrideTypeRole OverrideType = "role"
	OverrideTypeUser OverrideType = "user"
)

type Role struct {
	ID         uuid.UUID
	ServerID   uuid.UUID
	Name       string
	Permission authz.Permission
	Position   int
	CreatedAt  time.Time
}

type ChannelPermissionOverride struct {
	ChannelID  uuid.UUID
	TargetType OverrideType
	TargetID   uuid.UUID
	Allow      authz.Permission
	Deny       authz.Permission
}
