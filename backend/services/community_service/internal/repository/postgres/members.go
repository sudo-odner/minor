package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
)

func (repo *Repository) AddMember(
	ctx context.Context,
	serverID uuid.UUID,
	userID uuid.UUID,
	nickname *string,
) (*models.Member, error) {
	const op = "repository.postgres.AddMember"

	query := `
		INSERT INTO members (server_id, user_id, nickname, joined_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		RETURNING server_id, user_id, nickname, joined_at;
	`
	var member models.Member

	if err := repo.pool.QueryRow(ctx, query, serverID, userID, nickname).Scan(
		&member.ServerID,
		&member.UserID,
		&member.Nickname,
		&member.JoinedAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, models.ErrAlreadyExists
			}
			if pgErr.Code == "23503" {
				return nil, models.ErrNotFound
			}
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &member, nil
}

func (repo *Repository) GetServerMember(ctx context.Context, serverID, userID uuid.UUID) (*models.Member, error) {
	const op = "repository.postgres.GetServerMember"

	query := `
		SELECT server_id, user_id, nickname, joined_at
		FROM members
		WHERE server_id = $1 AND user_id = $2;
	`
	var member models.Member

	if err := repo.pool.QueryRow(ctx, query, serverID, userID).Scan(
		&member.ServerID,
		&member.UserID,
		&member.Nickname,
		&member.JoinedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	roles, err := repo.GetMemberRoles(ctx, serverID, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get member roles: %w", op, err)
	}
	if roles == nil {
		roles = make([]models.Role, 0)
	}
	member.Roles = roles

	return &member, nil
}

func (repo *Repository) GetServerMembers(ctx context.Context, serverID uuid.UUID) ([]models.Member, error) {
	const op = "repository.postgres.GetServerMembers"

	query := `
		SELECT server_id, user_id, nickname, joined_at
		FROM members
		WHERE server_id = $1;
	`
	rows, err := repo.pool.Query(ctx, query, serverID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var members []models.Member
	for rows.Next() {
		var member models.Member
		if err := rows.Scan(
			&member.ServerID,
			&member.UserID,
			&member.Nickname,
			&member.JoinedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: scan error: %w", op, err)
		}
		members = append(members, member)
	}

	// Fetch all member-role associations for this server in a single query (resolving N+1 query issue)
	rolesQuery := `
		SELECT mr.user_id, r.id, r.server_id, r.name, r.permissions, r.position, r.created_at
		FROM members_roles mr
		JOIN roles r ON mr.role_id = r.id
		WHERE mr.server_id = $1
		ORDER BY r.position DESC;
	`
	rolesRows, err := repo.pool.Query(ctx, rolesQuery, serverID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to query member roles: %w", op, err)
	}
	defer rolesRows.Close()

	rolesMap := make(map[uuid.UUID][]models.Role)
	for rolesRows.Next() {
		var userID uuid.UUID
		var role models.Role
		if err := rolesRows.Scan(
			&userID,
			&role.ID,
			&role.ServerID,
			&role.Name,
			&role.Permission,
			&role.Position,
			&role.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: scan role error: %w", op, err)
		}
		rolesMap[userID] = append(rolesMap[userID], role)
	}

	for i := range members {
		mRoles := rolesMap[members[i].UserID]
		if mRoles == nil {
			mRoles = make([]models.Role, 0)
		}
		members[i].Roles = mRoles
	}

	return members, nil
}

func (repo *Repository) RemoveMember(ctx context.Context, serverID, userID uuid.UUID) error {
	const op = "repository.postgres.RemoveMember"

	query := `DELETE FROM members WHERE server_id = $1 AND user_id = $2`
	res, err := repo.pool.Exec(ctx, query, serverID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return models.ErrNotFound
	}

	return nil
}

// Хз нужно ли возврощять объект пользователя ради 1 поля, нужно будет перепишу
func (repo *Repository) UpdateMemberNickname(
	ctx context.Context,
	serverID, userID uuid.UUID,
	nickname string,
) error {
	const op = "repository.postgres.UpdateMemberNickname"

	query := `
		UPDATE members 
		SET nickname = NULLIF($1, '')
		WHERE server_id = $2 AND user_id = $3;
	`
	res, err := repo.pool.Exec(ctx, query, nickname, serverID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return models.ErrNotFound
	}

	return nil
}

// defalut role? @everyone?
func (repo *Repository) AddRoleToMember(ctx context.Context, serverID, userID, roleID uuid.UUID) error {
	const op = "repository.postgres.AddRoleToMember"

	query := `
		INSERT INTO members_roles (server_id, user_id, role_id)
		VALUES ($1, $2, $3);
	`
	_, err := repo.pool.Exec(ctx, query, serverID, userID, roleID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				return models.ErrNotFound
			}
			if pgErr.Code == "23505" {
				return models.ErrAlreadyExists
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (repo *Repository) RemoveRoleFromMember(ctx context.Context, serverID, userID, roleID uuid.UUID) error {
	const op = "repository.postgres.RemoveRoleFromMember"

	query := `
		DELETE FROM members_roles WHERE server_id = $1 AND user_id = $2 AND role_id = $3;
	`
	res, err := repo.pool.Exec(ctx, query, serverID, userID, roleID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return models.ErrNotFound
	}

	return nil
}
