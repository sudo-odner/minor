package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		INSER INTO members (server_id, user_id, nickname, joined_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP);
	`
	var member models.Member

	if err := repo.pool.QueryRow(ctx, query, serverID, userID, nickname).Scan(
		&member.ServerID,
		&member.UserID,
		&member.Nickname,
		&member.JoinedAt,
	); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &member, nil
}

func (repo *Repository) GetServerMember(ctx context.Context, serverID, userID uuid.UUID) (*models.Member, error) {
	const op = "repository.postgres.GetServerMember"

	query := `
		SELECT server_id, user_id, nickname, joined_at
		FROM members
		WHERE server_id = $1 AND userID = $2;
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

	return &member, nil
}

func (repo *Repository) GetServerMembers(ctx context.Context, serverID uuid.UUID) ([]models.Member, error) {
	const op = "repository.postgres.GetServerMembers"

	query := `
		SELECT server_id, user_id, nickname, joined_at
		FROM memers
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

	return members, nil
}

func (repo *Repository) RemoveMember() {}

func (repo *Repository) UpdateMemberNickname() {}

func (repo *Repository) AddRoleToMember() {}

func (repo *Repository) RemoveRoleFromMember() {}
