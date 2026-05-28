package members

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
	"go.uber.org/zap"
)

type Repository interface {
	AddMember(ctx context.Context, serverID uuid.UUID, userID uuid.UUID, nickname *string) (*models.Member, error)
	GetServerMember(ctx context.Context, serverID, userID uuid.UUID) (*models.Member, error)
	GetServerMembers(ctx context.Context, serverID uuid.UUID) ([]models.Member, error)
	RemoveMember(ctx context.Context, serverID, userID uuid.UUID) error
	UpdateMemberNickname(ctx context.Context, serverID, userID uuid.UUID, nickname string) error
	AddRoleToMember(ctx context.Context, serverID, userID, roleID uuid.UUID) error
	RemoveRoleFromMember(ctx context.Context, serverID, userID, roleID uuid.UUID) error
}
type PermissionService interface {
	FetchServerPermissions(ctx context.Context, userID, serverID uuid.UUID) (authz.Permission, error)
}
type Service struct {
	log  *zap.Logger
	repo Repository
	perm PermissionService
}

func New(log *zap.Logger, repo Repository, permService PermissionService) *Service {
	return &Service{
		log:  log,
		repo: repo,
		perm: permService,
	}
}

func (s *Service) AddMember(ctx context.Context, serverID, userID uuid.UUID, nickname *string) (*models.Member, error) {
	const op = "service.member.AddMember"

	m, err := s.repo.AddMember(ctx, serverID, userID, nickname)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return m, nil
}

func (s *Service) GetServerMember(ctx context.Context, serverID, userID uuid.UUID) (*models.Member, error) {
	const op = "service.memver.GetServerMemver"

	m, err := s.repo.GetServerMember(ctx, serverID, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return m, nil
}

func (s *Service) GetServerMembers(ctx context.Context, serverID uuid.UUID) ([]models.Member, error) {
	const op = "service.member.GetServerMembers"

	ms, err := s.repo.GetServerMembers(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return ms, nil
}
