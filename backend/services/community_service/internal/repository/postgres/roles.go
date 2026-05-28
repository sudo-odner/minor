package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
)

func (repo *Repository) CreateRole(
	ctx context.Context,
	serverID uuid.UUID,
	name string,
	permission authz.Permission,
) (*models.Role, error) {
	const op = "repository.postgres.CreateRole"

	roleID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to generate uuid: %w", op, err)
	}

	query := `
		INSERT INTO roles (id, server_id, name, permission, positiion, created_at)
		VALUES ($1, $2, $3, $4, (SELECT COALESCE(MAX(position), 0) + 1 FROM roles WHERE server_id = $2), CURRENT_TIMESTAMP)
		RETURNING id, server_id, name, permission, position, created_at;
	`
	var role models.Role
	if err := repo.pool.QueryRow(ctx, query, roleID, serverID, name, permission).Scan(
		&role.ID,
		&role.ServerID,
		&role.Name,
		&role.Permission,
		&role.Position,
		&role.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &role, nil
}

func (repo *Repository) GetRole() {}

func (repo *Repository) GetServerRoles() {}

func (repo *Repository) GetMemberRoles() {}

func (repo *Repository) UpdateRole() {}

func (repo *Repository) DeleteRole() {}

func (repo *Repository) UpsertChannelOverride() {}

func (repo *Repository) DeleteChannelOverride() {}

func (repo *Repository) GetChannelOverrides() {}
