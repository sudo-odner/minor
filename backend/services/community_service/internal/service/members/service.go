package members

import (
	"context"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
	"go.uber.org/zap"
)

type Repository interface {
	AddMember(ctx context.Context, member *models.Member) (*models.Member, error)
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
	log         *zap.Logger
	repo        Repository
	permService PermissionService
}

func New(log *zap.Logger, repo Repository, permService PermissionService) *Service {
	return &Service{
		log:         log,
		repo:        repo,
		permService: permService,
	}
}
