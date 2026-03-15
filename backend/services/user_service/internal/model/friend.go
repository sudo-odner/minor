package model

import (
	"time"

	"github.com/google/uuid"
)

type FriendStatus string

const (
	FriendStatusPending  = "pending"
	FriendStatusAccepted = "accepted"
	FriendStatusDeny     = "deny"
	FriendStatusBlock    = "block"
)

type Friend struct {
	UserID   uuid.UUID
	FriendID uuid.UUID
	Status   FriendStatus
	CreateAt time.Time
}
