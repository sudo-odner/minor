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
			$1, 
			$2, 
			$3, 
			$4,
			$5, 
			-- Находим последнюю максимаьную позичию после которого будет вставка(по отношению к родителю)
			(
				SELECT COALESCE(MAX(position), 0) + 1
				FROM channels 
				WHERE server_id = $2 AND ((parent_id = $5) OR (parent_id IS NULL AND $5 IS NULL)
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
		OREDER BY 
			-- Сначала каналы бещ категорий
			(c.parent_id IS NULL AND c.type != 0) DESC,
			COALESCE(p.position, c.position) ASC,
			(c.parent_id IS NOT NULL)::int ASC,
			c.position ASC,
			c.created_at ASC;
		;
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

// TODO: продумать это чуть получше
// Если parendID == nil - не обновляем значение, если == uuid.Nil
func (repo *Repository) UpdateChannel(
	ctx context.Context,
	channelID, serverID uuid.UUID,
	name *string,
	parentID *uuid.UUID,
) (*models.Channel, error) {
	const op = "repository.postgres.UpdateChannel"

	var dbParentID any
	if parentID != nil {
		dbParentID = *parentID
	}

	query := `
		UPDATE channels 
		SET 
			name = $1, 
			parent_id = CASE 
				WHEN $2::uuid IS NULL THEN parent_id
				WHEN $2::uuid = '00000000-0000-0000-0000-000000000000'::uuid THEN NULL
				ELSE $2::uuid
			END
		RETURNING id, server_id, name, type, parent_id, position, created_at;
	`
	var channel models.Channel
	if err := repo.pool.QueryRow(ctx, query, name, dbParentID, channelID, serverID).Scan(
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
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &channel, nil
}

func (repo *Repository) DeleteChannel(ctx context.Context, channelID uuid.UUID) error {
	const op = "repository.postgres.DeleteChannel"

	query := `DELETE FROM channels WHERE id = $1`
	res, err := repo.pool.Exec(ctx, query, channelID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return models.ErrNotFound
	}

	return nil
}

func (repo *Repository) MoveChannel(
	ctx context.Context,
	serverID, channelID uuid.UUID,
	oldParentID, newParentID *uuid.UUID,
	oldPos, newPos int,
) error {
	const op = "repository.postgres.MoveChannel"

	query := `SELECT move_channel($1, $2, $3, $4, $5, $6)`
	_, err := repo.pool.Exec(ctx, query, serverID, channelID, oldParentID, newParentID, oldPos, newPos)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
