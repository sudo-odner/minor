package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
)

type Server struct {
	ID        uuid.UUID
	Name      string
	OwnerID   uuid.UUID
	AvatarURL string
	CreatedAt time.Time
}

type Member struct {
    UserID    uuid.UUID `json:"userId"`
    ServerID  uuid.UUID `json:"serverId"`
    Nickname  *string   `json:"nickname"` // Указатель для обработки NULL
    Username  string    `json:"username"`  // <-- ПРОВЕРЬ ЭТОТ ТЕГ
    AvatarURL string    `json:"avatarUrl"` // <-- И ЭТОТ
    JoinedAt  time.Time `json:"joinedAt"`
    Roles     []Role    `json:"roles"`
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
