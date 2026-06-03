package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrMessageNotFound  = errors.New("message not found")
	ErrChannelNotFound  = errors.New("channel not found")
	ErrInvalidChannel   = errors.New("invalid channel type")
)

type ChannelOwner string

const (
	ChannelOwnerCommunity ChannelOwner = "community"
	ChannelOwnerUser      ChannelOwner = "user"
)

type Message struct {
	ChannelID uuid.UUID
	MessageID uuid.UUID
	UserID    uuid.UUID
	Content   string
	ReplyTo   *uuid.UUID
	Username string
	CreatedAt time.Time
}
