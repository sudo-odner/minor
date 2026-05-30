package models

type ChatMessageCreated struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	AuthorID  string `json:"authod_id"`
	Content   string `json:"content"`
}
