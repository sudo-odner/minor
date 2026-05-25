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
	msgID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%s: failed generate uuid message: %w", op, err)
	}

	cqlMessageUUID, err := gocql.ParseUUID(msgID.String())
	if err != nil {
		return nil, fmt.Errorf("%s: invalid uuid conversion(messageID): %w", op, err)
	}
	cqlUserUUID, err := gocql.ParseUUID(userID.String())
	if err != nil {
		return nil, fmt.Errorf("%s: invalid uuid conversion(userID): %w", op, err)
	}
	cqlChannelUUID, err := gocql.ParseUUID(channelID.String())
	if err != nil {
		return nil, fmt.Errorf("%s: invalid uuid conversion(channelID): %w", op, err)
	}

	var cqlReplyTo any
	if replyTo != nil {
		parsedReply, err := gocql.ParseUUID(replyTo.String())
		if err != nil {
			return nil, fmt.Errorf("%s: invalid uuid conversion(replyTo): %w", op, err)
		}
		cqlReplyTo = parsedReply
	} else {
		cqlReplyTo = nil // Для Cassandra это означает записать NULL
	}

	query := `INSERT INTO messages (channel_id, message_id, author_id, content, reply_to, created_at) VALUES (?, ?, ?, ?, ?, ?);`
	err = r.session.Query(query, cqlChannelUUID, cqlMessageUUID, cqlUserUUID, content, cqlReplyTo, now).WithContext(ctx).Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.Message{
		ChannelID: channelID,
		MessageID: msgID,
		AuthorID:  userID,
		Content:   content,
		ReplyTo:   replyTo,
		CreatedAt: now,
	}, nil
}

func (r *Repository) GetMessages(ctx context.Context, channelID uuid.UUID, limit int, beforeID *uuid.UUID) ([]models.Message, error) {
	const op = "repository.cassandra.GetMessages"

	var query string
	var args []any

	cqlChannelUUID, err := gocql.ParseUUID(channelID.String())
	if err != nil {
		return nil, fmt.Errorf("%s: invalid uuid conversion(channelID): %w", op, err)
	}

	var cqlBeforeID any
	if beforeID != nil {
		parsedReply, err := gocql.ParseUUID(beforeID.String())
		if err != nil {
			return nil, fmt.Errorf("%s: invalid uuid conversion(replyTo): %w", op, err)
		}
		cqlBeforeID = parsedReply
	} else {
		cqlBeforeID = nil // Для Cassandra это означает записать NULL
	}

	if beforeID == nil {
		query = `SELECT channel_id, message_id, author_id, content, reply_to, created_at FROM messages WHERE channel_id = ? LIMIT ?;`
		args = []any{cqlChannelUUID, limit}
	} else {
		query = `SELECT channel_id, message_id, author_id, content, reply_to, created_at FROM messages WHERE channel_id = ? AND message_id < ? LIMIT ?;`
		args = []any{cqlChannelUUID, cqlBeforeID, limit}
	}

	iter := r.session.Query(query, args...).WithContext(ctx).Iter()

	messages := make([]models.Message, 0, limit)
	var (
		cqlChan   gocql.UUID
		cqlMsg    gocql.UUID
		cqlAuth   gocql.UUID
		content   string
		cqlReply  *gocql.UUID
		createdAt time.Time
	)

	for iter.Scan(&cqlChan, &cqlMsg, &cqlAuth, &content, &cqlReply, &createdAt) {
		goChanID, _ := uuid.Parse(cqlChan.String())
		goMsgID, _ := uuid.Parse(cqlMsg.String())
		goAuthID, _ := uuid.Parse(cqlAuth.String())

		var goReplyTo *uuid.UUID
		if cqlReply != nil {
			parsedReply, err := uuid.Parse(cqlReply.String())
			if err == nil {
				goReplyTo = &parsedReply
			}
		}

		// Собираем бизнес-модель
		messages = append(messages, models.Message{
			ChannelID: goChanID,
			MessageID: goMsgID,
			AuthorID:  goAuthID,
			Content:   content,
			ReplyTo:   goReplyTo,
			CreatedAt: createdAt,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("%s: iterator error: %w", op, err)
	}

	return messages, nil
}

func (r *Repository) GetMessage(ctx context.Context, channelID, messageID uuid.UUID) (*models.Message, error) {
	const op = "repository.cassandra.GetMessage"

	cqlMessageUUID, err := gocql.ParseUUID(messageID.String())
	if err != nil {
		return nil, fmt.Errorf("%s: invalid uuid conversion(messageID): %w", op, err)
	}
	cqlChannelUUID, err := gocql.ParseUUID(channelID.String())
	if err != nil {
		return nil, fmt.Errorf("%s: invalid uuid conversion(channelID): %w", op, err)
	}

	var (
		cqlChan   gocql.UUID
		cqlMsg    gocql.UUID
		cqlAuth   gocql.UUID
		content   string
		cqlReply  *gocql.UUID
		createdAt time.Time
	)

	query := `SELECT channel_id, message_id, author_id, content, reply_to, created_at 
			  FROM messages WHERE channel_id = ? AND message_id = ?;`

	err = r.session.Query(query, cqlChannelUUID, cqlMessageUUID).
		WithContext(ctx).
		Scan(&cqlChan, &cqlMsg, &cqlAuth, &content, &cqlReply, &createdAt)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, fmt.Errorf("%s: %w", op, models.ErrMessageNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	goChanID, _ := uuid.Parse(cqlChan.String())
	goMsgID, _ := uuid.Parse(cqlMsg.String())
	goAuthID, _ := uuid.Parse(cqlAuth.String())

	var goReplyTo *uuid.UUID
	if cqlReply != nil {
		parsedReply, err := uuid.Parse(cqlReply.String())
		if err == nil {
			goReplyTo = &parsedReply
		}
	}
	return &models.Message{
		ChannelID: goChanID,
		MessageID: goMsgID,
		AuthorID:  goAuthID,
		Content:   content,
		ReplyTo:   goReplyTo,
		CreatedAt: createdAt,
	}, nil
}

func (r *Repository) DeleteMessage(ctx context.Context, channelID, messageID uuid.UUID) error {
	const op = "repository.cassandra.DeleteMessage"

	cqlMessageUUID, err := gocql.ParseUUID(messageID.String())
	if err != nil {
		return fmt.Errorf("%s: invalid uuid conversion(messageID): %w", op, err)
	}
	cqlChannelUUID, err := gocql.ParseUUID(channelID.String())
	if err != nil {
		return fmt.Errorf("%s: invalid uuid conversion(channelID): %w", op, err)
	}

	query := `DELETE FROM messages WHERE channel_id = ? AND message_id = ?;`
	if err := r.session.Query(query, cqlChannelUUID, cqlMessageUUID).WithContext(ctx).Exec(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
