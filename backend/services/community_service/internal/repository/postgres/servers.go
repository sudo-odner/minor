package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
)

func (repo *Repository) CreateServer(
	ctx context.Context,
	name string,
	ownerID uuid.UUID,
	avatarURL string,
) (*models.Server, error) {
	const op = "repository.postgres.CreateServer"

	newID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	query := `
		INSERT INTO servers (id, name, owner_id, avatar_url, created_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
		RETURNING id, name, owner_id, avatar_url, created_at;
	`
	var server models.Server

	if err := repo.pool.QueryRow(ctx, query, newID, name, ownerID, avatarURL).Scan(
		&server.ID,
		&server.Name,
		&server.OwnerID,
		&server.AvatarURL,
		&server.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &server, nil
}

func (repo *Repository) GetServer(ctx context.Context, serverID uuid.UUID) (*models.Server, error) {
	const op = "repository.postgres.GetServer"

	query := `
		SELECT id, name, owner_id, avatar_url, created_at
		FROM servers
		WHERE id = $1;
	`
	var server models.Server

	if err := repo.pool.QueryRow(ctx, query, serverID).Scan(
		&server.ID,
		&server.Name,
		&server.OwnerID,
		&server.AvatarURL,
		&server.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &server, nil
}

func (repo *Repository) UpdateServer(
	ctx context.Context,
	serverID uuid.UUID,
	name *string,
	avatarURL *string,
) (*models.Server, error) {
	const op = "repository.postgres.UpdateServer"

	query := `
		UPDATE servers 
		SET
			name = COALESCE($1, name),
			avatar_url = COALESCE($2, avatar_url)
		WHERE id = $3
		RETURNING id, name, owner_id, avatar_url, created_at;
	`
	var server models.Server

	if err := repo.pool.QueryRow(ctx, query, name, avatarURL, serverID).Scan(
		&server.ID,
		&server.Name,
		&server.OwnerID,
		&server.AvatarURL,
		&server.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &server, nil
}

func (repo *Repository) DeleteServer(ctx context.Context, serverID uuid.UUID) error {
	const op = "repository.postgres.DeleteServer"

	query := `DELETE FROM servers WHERE id = $1;`
	res, err := repo.pool.Exec(ctx, query, serverID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}
