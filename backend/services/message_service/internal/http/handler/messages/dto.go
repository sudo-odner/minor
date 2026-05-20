package messages

import "time"

type Message struct {
	ChannelID string    `json:"channel_id"`
	MessageID string    `json:"message_id"`
	AuthorID  string    `json:"author_id"`
	Content   string    `json:"content"`
	ReplyTo   string    `json:"reply_to"`
	CreateAt  time.Time `json:"create_at"`
}

type ReqSaveMessage struct {
	ChannelID string `json:"channel_id" validate:"required,uuid"`
	Content   string `json:"content" validate:"required"`
	ReplyTo   string `json:"reply_to" validate:"omitempty, uuid"`
}

type ResGetMessages struct {
	Messages []Message `json:"messages"`
}
