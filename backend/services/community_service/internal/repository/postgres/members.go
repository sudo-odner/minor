package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
)

func (repo *Repository) AddMember(
	ctx context.Context,
	serverID uuid.UUID,
	userID uuid.UUID,
	nickname *string,
) (*models.Members, error) {
	const op = "repository.postgres.AddMember"

	query := `
		INSER INTO members (server_id, user_id, nickname, joined_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP);
	`
	var member models.Members

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

func (repo *Repository) GetServerMember() {}

func (repo *Repository) GetServerMembers() {}

func (repo *Repository) RemoveMember() {}

func (repo *Repository) UpdateMemberNickname() {}

func (repo *Repository) AddRoleToMember() {}

func (repo *Repository) RemoveRoleFromMember() {}
