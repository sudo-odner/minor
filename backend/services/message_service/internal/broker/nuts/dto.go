package nuts

import "time"

const (
	SubjectMessageCreated = "chat.message.created"
	SubjectMessageDeleted = "chat.message.deleted"
)

type MessageCreatedEvent struct {
	MessageID string    `json:"message_id"`
	ChannelID string    `json:"channel_id"`
	AuthorID  string    `json:"author_id"`
	Content   string    `json:"content"`
	ReplyTo   *string   `json:"reply_to,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type MessageDeletedEvent struct {
	MessageID string `json:"message_id"`
	ChannelID string `json:"channel_id"`
}
