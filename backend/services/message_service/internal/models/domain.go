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
	ChannelOwnerDM        ChannelOwner = "dm"
)

// Message represents a single chat message sent within a channel.
type Message struct {
	// ChannelID identifies the channel where the message was posted.
	ChannelID uuid.UUID `json:"channel_id" db:"channel_id"`

	// MessageID is the unique identifier of the message (usually UUIDv7).
	MessageID uuid.UUID `json:"message_id" db:"message_id"`

	// UserID is the author's unique identifier.
	UserID uuid.UUID `json:"user_id" db:"user_id"`

	// Content contains the raw text body of the message.
	Content string `json:"content" db:"content"`

	// ReplyTo points to the parent message ID if this is a reply; nil otherwise.
	ReplyTo *uuid.UUID `json:"reply_to,omitempty" db:"reply_to"`

	// CreatedAt is the UTC timestamp when the message was created.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
