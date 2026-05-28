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

func (repo *Repository) UpdateRole(
	ctx context.Context,
	roleID uuid.UUID,
	name *string,
	permissions *authz.Permission,
) (*models.Role, error) {
	const op = "repository.postgres.UpdateRole"

	query := `
		UPDATE roles
		SET
			name = COALESCE($1, name),
			permissions = COALESCE($2, permission)
		WHERE id = $3
		RETURNING id, server_id, name, permission, position, created_at;
	`
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

func (repo *Repository) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	const op = "repository.postgres.DeleteRole"

	query := `DELTE FROM roles WHERE id = $1`
	res, err := repo.pool.Exec(ctx, query, roleID)
	if err != nil {
		return fmt.Errorf("%s: delete role falied: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return models.ErrNotFound
	}

	return nil
}

func (repo *Repository) ReplaceChannelPermissionOverrides(
	ctx context.Context,
	channelID uuid.UUID,
	overrides []models.ChannelPermissionOverride,
) error {
	const op = "repository.postgres.ReplaceChannelPermissionOverrides"

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: begin tx falied: %w", op, err)
	}
	defer tx.Rollback(ctx)

	deleteQuery := `DELETE FROM channel_permission_overrides WHERE channel_id = $1`
	_, err = tx.Exec(ctx, deleteQuery, channelID)
	if err != nil {
		return fmt.Errorf("%s: clear overrieds failed: %w", op, err)
	}

	if len(overrides) <= 0 {
		return tx.Commit(ctx)
	}

	insertQuery := `
		INSERT INTO channel_permission_overrides (channel_id, target_type, target_id, allow, deny)
		VALUES ($1, $2, $3, $4, $5);
	`
	for _, ov := range overrides {
		_, err = tx.Exec(ctx, insertQuery, channelID, ov.TargetType, ov.TargetID, ov.Allow, ov.Deny)
		if err != nil {
			return fmt.Errorf("%s: insert override falied: %w", op, err)
		}
	}

	return tx.Commit(ctx)
}

func (repo *Repository) GetChannelOverrides(ctx context.Context, channelID uuid.UUID) ([]models.ChannelPermissionOverride, error) {
	const op = "repository.postgres.GetChannelOverrides"

	query := `SELECT channel_id, target_type, target_id, allow, deny FROM channel_permission_overrides WHERE cahnnel_id = $1`
	rows, err := repo.pool.Query(ctx, query, channelID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var overrides []models.ChannelPermissionOverride
	for rows.Next() {
		var ov models.ChannelPermissionOverride
		if err := rows.Scan(&ov.ChannelID, &ov.TargetType, &ov.TargetID, &ov.Allow, &ov.Deny); err != nil {
			return nil, fmt.Errorf("%s: scan error: %w", op, err)
		}
		overrides = append(overrides, ov)
	}

	return overrides, nil
}
