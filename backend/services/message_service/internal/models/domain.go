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

type ChannelType string

const (
	ChannelTypeGuild ChannelType = "guild"
	ChannelTypeDM    ChannelType = "dm"
)

type Message struct {
	ChannelID uuid.UUID
	MessageID uuid.UUID
	AuthorID  uuid.UUID
	Content   string
	ReplyTo   *uuid.UUID
	CreatedAt time.Time
}
