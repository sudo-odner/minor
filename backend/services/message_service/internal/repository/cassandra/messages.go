package cassandra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/models"
)

// TODO: Write SaveBatchMessage for asinc save

// SaveMessage creates and persists a new message in Cassandra for the specified channel and user.
// The replyTo parameter is optional; pass nil if the message is not a reply (not zero value uuid).
func (r *Repository) SaveMessage(ctx context.Context, userID, channelID uuid.UUID, content string, replyTo *uuid.UUID) (*models.Message, error) {
	const op = "repository.cassandra.SaveMessage"

	now := time.Now().UTC()
	messageID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%s: failed generate uuid message: %w", op, err)
	}

	query := `
	INSERT INTO messages (channel_id, message_id, user_id, content, reply_to, created_at) 
	VALUES (?, ?, ?, ?, ?, ?);`
	err = r.session.Query(
		query,
		gocql.UUID(channelID),
		gocql.UUID(messageID),
		gocql.UUID(userID),
		content,
		(*gocql.UUID)(replyTo),
		now,
	).WithContext(ctx).Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.Message{
		ChannelID: channelID,
		MessageID: messageID,
		UserID:    userID,
		Content:   content,
		ReplyTo:   replyTo,
		CreatedAt: now,
	}, nil
}

// GetMessages retrieves up to limit messages from the specified channel in reverse chronological order.
// If beforeID is provided, it returns messages created prior to that message ID for pagination.
func (r *Repository) GetMessages(ctx context.Context, channelID uuid.UUID, limit int, beforeID *uuid.UUID) ([]models.Message, error) {
	const op = "repository.cassandra.GetMessages"

	var query string
	var args []any

	if beforeID == nil {
		query = `
		SELECT channel_id, message_id, user_id, content, reply_to, created_at 
		FROM messages 
		WHERE channel_id = ? 
		LIMIT ?;`
		args = []any{gocql.UUID(channelID), limit}
	} else {
		query = `
		SELECT channel_id, message_id, user_id, content, reply_to, created_at 
		FROM messages 
		WHERE channel_id = ? AND message_id < ? 
		LIMIT ?;`
		args = []any{gocql.UUID(channelID), (*gocql.UUID)(beforeID), limit}
	}

	iter := r.session.Query(query, args...).WithContext(ctx).Iter()
	messages := make([]models.Message, 0, limit)
	var (
		mChannelID, mMessageID, mUserID gocql.UUID
		mContent                        string
		mReplyTo                        *gocql.UUID
		mCreatedAt                      time.Time
	)
	for iter.Scan(&mChannelID, &mMessageID, &mUserID, &mContent, &mReplyTo, &mCreatedAt) {
		messages = append(messages, models.Message{
			ChannelID: uuid.UUID(mChannelID),
			MessageID: uuid.UUID(mMessageID),
			UserID:    uuid.UUID(mUserID),
			Content:   mContent,
			ReplyTo:   (*uuid.UUID)(mReplyTo),
			CreatedAt: mCreatedAt,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("%s: iterator error: %w", op, err)
	}

	return messages, nil
}

// GetMessage get message from the specified channel by messageID.
func (r *Repository) GetMessage(ctx context.Context, channelID, messageID uuid.UUID) (*models.Message, error) {
	const op = "repository.cassandra.GetMessage"

	var (
		mChannelID, mMessageID, mUserID gocql.UUID
		mContent                        string
		mReplyTo                        *gocql.UUID
		mCreatedAt                      time.Time
	)

	err := r.session.Query(
		`
			SELECT channel_id, message_id, user_id, content, reply_to, created_at 
			FROM messages 
			WHERE channel_id = ? AND message_id = ?;
		`,
		gocql.UUID(messageID),
		gocql.UUID(channelID),
	).WithContext(ctx).Scan(&mChannelID, &mMessageID, &mUserID, &mContent, &mReplyTo, &mCreatedAt)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, fmt.Errorf("%s: %w", op, models.ErrMessageNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.Message{
		ChannelID: uuid.UUID(mChannelID),
		MessageID: uuid.UUID(mMessageID),
		UserID:    uuid.UUID(mUserID),
		Content:   mContent,
		ReplyTo:   (*uuid.UUID)(mReplyTo),
		CreatedAt: mCreatedAt,
	}, nil
}

// DeleteMessage delete message from the specified channel by messageID.
func (r *Repository) DeleteMessage(ctx context.Context, channelID, messageID uuid.UUID) error {
	const op = "repository.cassandra.DeleteMessage"

	query := `DELETE FROM messages WHERE channel_id = ? AND message_id = ?;`
	if err := r.session.Query(query, gocql.UUID(channelID), gocql.UUID(messageID)).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
