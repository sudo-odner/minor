package models

import "time"

type Message struct {
	ChannelID []byte
	MessageID []byte
	AuthorID  []byte
	Content   string
	ReplyTo   []byte
	CreatedAt time.Time
}
