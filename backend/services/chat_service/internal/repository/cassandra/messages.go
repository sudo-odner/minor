package cassandra

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/chat_service/internal/models"
)

func (r *Repository) SaveMessage(ctx context.Context, userID, channelID uuid.UUID, content string, replyTo *uuid.UUID) (models.Message, error) {
	const op = "repository.cassandra.SaveMessage"

	msgID, err := uuid.NewV7()
	if err != nil {
		return models.Message{}, fmt.Errorf("%s: failed generate uuid message: %w", op, err)
	}
	now := time.Now().UTC()

	query := `INSERT INTO message (channel_id, message_id, author_id, content, reply_to, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	err = r.session.Query(query, channelID, msgID, userID, content, replyTo, now).WithContext(ctx).Exec()
	if err != nil {
		return models.Message{}, fmt.Errorf("%s: %w", op, err)
	}

	return models.Message{
		ChannelID: channelID,
		MessageID: msgID,
		AuthorID:  userID,
		Content:   content,
		ReplyTo:   replyTo,
		CreatedAt: now,
	}, nil
}

func (r *Repository) GetMessages(ctx context.Context, channelID uuid.UUID, limit int, beforeID *uuid.UUID) ([]models.Message, error) {
	const op = "repositiry.cassandra.GetMessages"

	var query string
	var args []any

	if beforeID == nil {
		query = `SELECT channel_id, message_id, author_id, content, reply_to, created_at FROM message WHERE channel_id = ? LIMIT ?`
		args = []any{channelID, limit}
	} else {
		query = `SELECT channel_id, message_id, author_id, content, reply_to, created_at FROM message WHERE channel_id = ? AND message_id < ? LIMIT ?`
		args = []any{channelID, beforeID, limit}
	}

	iter := r.session.Query(query, args...).WithContext(ctx).Iter()

	messages := make([]models.Message, 0, limit)
	var m models.Message

	for iter.Scan(&m.ChannelID, &m.MessageID, &m.AuthorID, &m.Content, &m.ReplyTo, &m.CreatedAt) {
		messages = append(messages, m)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return messages, nil
}

func (r *Repository) DeleteMessage(ctx context.Context, channelID, messageID uuid.UUID) error {
	const op = "repository.cassandra.DeleteMessage"

	query := `DELETE FROM messages WHERE channel_id = ? AND message_id = ?`
	if err := r.session.Query(query, channelID, messageID).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
