package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
)

func (repo *Repository) CreateServerWithDefaultSetup(
	ctx context.Context,
	name string,
	ownerID uuid.UUID,
	avatarURL string,
) (*models.Server, error) {
	const op = "repository.postgres.CreateServer"

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: begin tx failed: %w", op, err)
	}
	defer tx.Rollback(ctx)

	serverID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%s: uuid generation failed: %w", op, err)
	}

	queryServer := `
		INSERT INTO servers (id, name, owner_id, avatar_url, created_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP);
	`
	_, err = tx.Exec(ctx, queryServer, serverID, name, ownerID, avatarURL)
	if err != nil {
		return nil, fmt.Errorf("%s: insert server failed: %w", op, err)
	}

	// TODO: Поменять маску прав на чтение/запись
	queryRole := `
		INSERT INTO roles (id, server_id, name, permissions, position, created_at)
		VALUES ($1, $1, $2, $3, $4, CURRENT_TIMESTAMP);
	`
	_, err = tx.Exec(ctx, queryRole, serverID, "@everyone", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("%s: insert default role failed: %w", op, err)
	}

	queryMember := `
		INSERT INTO members (server_id, user_id, nickname, joined_at)
		VALUES ($1, $2, NULL, CURRENT_TIMESTAMP);
	`
	_, err = tx.Exec(ctx, queryMember, serverID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("%s: add owner as member failed: %w", op, err)
	}

	// TODO: Привязывать пользователей к @everyone? (Скорее нет, чем да)

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: commit failed: %w", op, err)
	}

	return &models.Server{
		ID:        serverID,
		Name:      name,
		OwnerID:   ownerID,
		AvatarURL: avatarURL,
	}, nil
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
	name string,
	avatarURL string,
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

func (repo *Repository) GetUserServers(ctx context.Context, userID uuid.UUID) ([]models.Server, error) {
	const op = "repository.postgres.GetUserServers"

	query := `
		SELECT s.id, s.name, s.owner_id, s.avatar_url, s.created_at
		FROM servers s
		JOIN members m ON s.id = m.server_id
		WHERE m.user_id = $1;
	`
	rows, err := repo.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}
	defer rows.Close()

	var servers []models.Server
	for rows.Next() {
		var server models.Server
		if err := rows.Scan(
			&server.ID,
			&server.Name,
			&server.OwnerID,
			&server.AvatarURL,
			&server.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: scan failed: %w", op, err)
		}
		servers = append(servers, server)
	}

	return servers, nil
}
