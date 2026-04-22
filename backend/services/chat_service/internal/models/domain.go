package models

import (
	"time"

	"github.com/google/uuid"
)

type NewMessage struct {
	ChannelID uuid.UUID
	AuthorID  uuid.UUID
	Content   string
	ReplyTo   uuid.UUID
}

type Message struct {
	ChannelID uuid.UUID
	MessageID uuid.UUID
	AuthorID  uuid.UUID
	Content   string
	ReplyTo   uuid.UUID
	CreatedAt time.Time
}
