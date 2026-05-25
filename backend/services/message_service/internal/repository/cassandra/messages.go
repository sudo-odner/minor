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

// TODO: Продумать replyTo *uuid.UUID
func (r *Repository) SaveMessage(ctx context.Context, userID, channelID uuid.UUID, content string, replyTo *uuid.UUID) (*models.Message, error) {
	const op = "repository.cassandra.SaveMessage"

	now := time.Now().UTC()
	messageID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%s: failed generate uuid message: %w", op, err)
	}

	cqlMessageID := gocql.UUID(messageID)
	cqlUserID := gocql.UUID(userID)
	cqlChannelID := gocql.UUID(channelID)

	var cqlReplyTo any = nil
	if replyTo != nil {
		cqlReplyTo = gocql.UUID(*replyTo)
	}

	query := `
	INSERT INTO messages (channel_id, message_id, user_id, content, reply_to, created_at) 
	VALUES (?, ?, ?, ?, ?, ?);`
	err = r.session.Query(
		query,
		cqlChannelID,
		cqlMessageID,
		cqlUserID,
		content,
		cqlReplyTo,
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

func (r *Repository) GetMessages(ctx context.Context, channelID uuid.UUID, limit int, beforeID *uuid.UUID) ([]models.Message, error) {
	const op = "repository.cassandra.GetMessages"

	var query string
	var args []any

	cqlChannelID := gocql.UUID(channelID)
	var cqlBeforeID any = nil
	if beforeID != nil {
		cqlBeforeID = gocql.UUID(*beforeID)
	}

	if beforeID == nil {
		query = `
		SELECT channel_id, message_id, user_id, content, reply_to, created_at 
		FROM messages 
		WHERE channel_id = ? 
		LIMIT ?;`
		args = []any{cqlChannelID, limit}
	} else {
		query = `
		SELECT channel_id, message_id, user_id, content, reply_to, created_at 
		FROM messages 
		WHERE channel_id = ? AND message_id < ? 
		LIMIT ?;`
		args = []any{cqlChannelID, cqlBeforeID, limit}
	}

	iter := r.session.Query(query, args...).WithContext(ctx).Iter()
	messages := make([]models.Message, 0, limit)
	var (
		cqlIterChannelID gocql.UUID
		cqlIterMessageID gocql.UUID
		cqlIterUserID    gocql.UUID
		cqlIterContent   string
		cqlIterReplyTo   *gocql.UUID
		cqlIterCreatedAt time.Time
	)
	for iter.Scan(&cqlChannelID, &cqlIterMessageID, &cqlIterUserID, &cqlIterContent, &cqlIterReplyTo, &cqlIterCreatedAt) {
		iterChannelID := uuid.UUID(cqlIterChannelID)
		iterMessageID := uuid.UUID(cqlIterMessageID)
		iterUserID := uuid.UUID(cqlIterUserID)

		var iterReplyTo *uuid.UUID
		if cqlIterReplyTo != nil {
			parsedUUID := uuid.UUID(*cqlIterReplyTo)
			iterReplyTo = &parsedUUID
		}

		messages = append(messages, models.Message{
			ChannelID: iterChannelID,
			MessageID: iterMessageID,
			UserID:    iterUserID,
			Content:   cqlIterContent,
			ReplyTo:   iterReplyTo,
			CreatedAt: cqlIterCreatedAt,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("%s: iterator error: %w", op, err)
	}

	return messages, nil
}

func (r *Repository) GetMessage(ctx context.Context, channelID, messageID uuid.UUID) (*models.Message, error) {
	const op = "repository.cassandra.GetMessage"

	cqlMessageID := gocql.UUID(messageID)
	cqlChannelID := gocql.UUID(channelID)
	var (
		cqlIterChannelID gocql.UUID
		cqlIterMessageID gocql.UUID
		cqlIterUserID    gocql.UUID
		cqlIterContent   string
		cqlIterReplyTo   *gocql.UUID
		cqlIterCreatedAt time.Time
	)

	query := `
	SELECT channel_id, message_id, user_id, content, reply_to, created_at 
	FROM messages 
	WHERE channel_id = ? AND message_id = ?;`

	err := r.session.Query(
		query,
		cqlChannelID,
		cqlMessageID,
	).WithContext(ctx).Scan(
		&cqlIterChannelID,
		&cqlIterMessageID,
		&cqlIterUserID,
		&cqlIterContent,
		&cqlIterReplyTo,
		&cqlIterCreatedAt,
	)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, fmt.Errorf("%s: %w", op, models.ErrMessageNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	iterChannelID := uuid.UUID(cqlIterChannelID)
	iterMessageID := uuid.UUID(cqlIterMessageID)
	iterUserID := uuid.UUID(cqlIterUserID)

	var iterReplyTo *uuid.UUID
	if cqlIterReplyTo != nil {
		parsedUUID := uuid.UUID(*cqlIterReplyTo)
		iterReplyTo = &parsedUUID
	}

	return &models.Message{
		ChannelID: iterChannelID,
		MessageID: iterMessageID,
		UserID:    iterUserID,
		Content:   cqlIterContent,
		ReplyTo:   iterReplyTo,
		CreatedAt: cqlIterCreatedAt,
	}, nil
}

func (r *Repository) DeleteMessage(ctx context.Context, channelID, messageID uuid.UUID) error {
	const op = "repository.cassandra.DeleteMessage"

	cqlChannelID := gocql.UUID(channelID)
	cqlMessageID := gocql.UUID(messageID)

	query := `DELETE FROM messages WHERE channel_id = ? AND message_id = ?;`
	if err := r.session.Query(query, cqlChannelID, cqlMessageID).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
