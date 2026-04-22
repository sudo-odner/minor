package models

import "time"

type NewMessage struct {
	ChannelID []byte
	AuthorID  []byte
	Content   string
	ReplyTo   []byte
}

type Message struct {
	ChannelID []byte
	MessageID []byte
	AuthorID  []byte
	Content   string
	ReplyTo   []byte
	CreatedAt time.Time
}
