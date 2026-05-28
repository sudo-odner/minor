package roles

import (
	"context"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
	"go.uber.org/zap"
)

type Repository interface {
	CreateRole(ctx context.Context, serverID uuid.UUID, name string, permissions authz.Permission) (*models.Role, error)
	GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error)
	GetServerRoles(ctx context.Context, serverID uuid.UUID) ([]models.Role, error)
	GetMemberRoles(ctx context.Context, serverID, userID uuid.UUID) ([]models.Role, error)
	UpdateRole(ctx context.Context, roleID uuid.UUID, name *string, permissions *authz.Permission) (*models.Role, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID) error
	ReplaceChannelPermissionOverrides(ctx context.Context, channelID uuid.UUID, overrides []models.ChannelPermissionOverride) error
}

type ServiceServer interface {
	GetServer(ctx context.Context, serverID uuid.UUID) (*models.Server, error)
}

type Service struct {
	log  *zap.Logger
	repo Repository
}

func New(log *zap.Logger, repo Repository) *Service {
	return &Service{
		log:  log,
		repo: repo,
	}
}
