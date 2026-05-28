package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
)

func (repo *Repository) CreateChannel(
	ctx context.Context,
	serverID uuid.UUID,
	name string,
	typeChannel models.ChannelType,
	parentID *uuid.UUID,
) (*models.Channel, error) {
	const op = "repository.postgres.CreateChannel"

	newUUID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%s: uuid generation failed: %w", op, err)
	}

	query := `
		INSERT INTO channels (id, server_id, name, type, parent_id, position, created_at)
		VALUES (
			$1, $2, $3, $4, $5, (
				SELECT COALESCE(MAX(position), 0) + 1
				FROM channels WHERE server_id = $2
			),
			CURRENT_TIMESTAMP
        )
		RETURNING id, server_id, name, type, parent_id, position, created_at;
	`
	var channel models.Channel

	if err := repo.pool.QueryRow(ctx, query, newUUID, serverID, name, typeChannel, parentID).Scan(
		&channel.ID,
		&channel.ServerID,
		&channel.Name,
		&channel.Type,
		&channel.ParentID,
		&channel.Position,
		&channel.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("%s: insert channel falied: %w", op, err)
	}

	return &channel, nil
}

func (repo *Repository) GetChannel(ctx context.Context, channelID uuid.UUID) (*models.Channel, error) {
	const op = "repository.postgres.GetChannel"

	query := `
		SELECT id, server_id, name, type, parent_id, position, created_at
		FROM channels
		WHERE id = $1;
	`
	var channel models.Channel

	if err := repo.pool.QueryRow(ctx, query, channelID).Scan(
		&channel.ID,
		&channel.ServerID,
		&channel.Name,
		&channel.Type,
		&channel.ParentID,
		&channel.Position,
		&channel.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("%s: find channel failed: %w", op, err)
	}

	return &channel, nil
}

func (repo *Repository) GetServerChannels(ctx context.Context, serverID uuid.UUID) ([]models.Channel, error) {
	const op = "repository.postgres.GetChannels"

	query := `
		SELECT id, server_id, name, type, parent_id, position, created_at
		FROM channels
		WHERE server_id = $1
		OREDER BY position ASC, created_at ASC;
	`
	rows, err := repo.pool.Query(ctx, query, serverID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var channels []models.Channel
	if rows.Next() {
		var channel models.Channel

		if err := rows.Scan(
			&channel.ID,
			&channel.ServerID,
			&channel.Name,
			&channel.Type,
			&channel.ParentID,
			&channel.Position,
			&channel.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: scan error: %w", op, err)
		}
		channels = append(channels, channel)
	}
	return nil, fmt.Errorf("%s: find channel failed: %w", op, err)
}

func (repo *Repository) UpdateChannel() {}

func (repo *Repository) DeleteChannel() {}

func (repo *Repository) MoveChannel() {}
