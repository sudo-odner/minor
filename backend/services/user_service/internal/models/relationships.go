package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RelationshipStatus int16

const (
	StatusFriends         RelationshipStatus = 1 // Друзья
	StatusRequestSent     RelationshipStatus = 2 // Заявка отправлена
	StatusRequestReceived RelationshipStatus = 3 // Ожидание подтверждения
	StatusBlocked         RelationshipStatus = 4 // В черном списке
)

func (s RelationshipStatus) String() string {
	switch s {
	case StatusFriends:
		return "friends"
	case StatusRequestSent:
		return "request_sent"
	case StatusRequestReceived:
		return "request_received"
	case StatusBlocked:
		return "blocked"
	default:
		return fmt.Sprintf("unknown_status_%d", s)
	}
}

type Relationship struct {
	UserID    uuid.UUID
	TargetID  uuid.UUID
	Status    RelationshipStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RelationshipPreview struct {
	UserID    uuid.UUID
	Username  string
	AvatarURL string
	Status    RelationshipStatus
	IsOnline  bool
}
