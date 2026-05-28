package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (repo *Repository) GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	const op = "repository.postgres.GetRole"

	query := `SELECT id, server_id, name, permission, position, created_at FROM roles WHERE id = $1;`
	var role models.Role

	if err := repo.pool.QueryRow(ctx, query, roleID).Scan(
		&role.ID,
		&role.ServerID,
		&role.Name,
		&role.Permission,
		&role.Position,
		&role.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &role, nil
}

func (repo *Repository) GetServerRoles(ctx context.Context, serverID uuid.UUID) ([]models.Role, error) {
	const op = "repository.postgres.GetServerRoles"

	query := `SELECT id, server_id, name, permission, position, created_at FROM roles WHERE server_id = $1 ORDER BY position DESC;`
	rows, err := repo.pool.Query(ctx, query, serverID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(
			&role.ID,
			&role.ServerID,
			&role.Name,
			&role.Permission,
			&role.Position,
			&role.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: scan error: %w", op, err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// TODO: Мб переделать в 1 метод кторый получает всю инфу о пользователе на сервере?
func (repo *Repository) GetMemberRoles(ctx context.Context, serverID, userID uuid.UUID) ([]models.Role, error) {
	const op = "repository.postgres.GetMemberRoles"

	query := `
		SELECT id, server_id, name, permission, position, created_at 
		FROM roles r
		JOIN members_roles mr ON r.id = mr.role_id
		WHERE mr.server_id = $1 AND mr.user_id = $2 
		ORDER BY position DESC;
	`
	rows, err := repo.pool.Query(ctx, query, serverID, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(
			&role.ID,
			&role.ServerID,
			&role.Name,
			&role.Permission,
			&role.Position,
			&role.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: scan error: %w", op, err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (repo *Repository) UpdateRole() {}

func (repo *Repository) DeleteRole() {}

func (repo *Repository) UpsertChannelOverride() {}

func (repo *Repository) DeleteChannelOverride() {}

func (repo *Repository) GetChannelOverrides() {}
