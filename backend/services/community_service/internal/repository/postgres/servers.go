package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
		RETURNING id, name, owner_id, avatar_url, created_at
	`
	var newServer models.Server

	if err := repo.pool.QueryRow(ctx, query, newID, name, ownerID, avatarURL).Scan(
		&newServer.ID,
		&newServer.Name,
		&newServer.OwnerID,
		&newServer.AvatarURL,
		&newServer.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &newServer, nil
}

func (repo *Repository) GetServer() {}

func (repo *Repository) UpdateServer() {}

func (repo *Repository) DeleteServer() {}
