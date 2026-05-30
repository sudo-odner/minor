package roles

import (
	"context"
	"fmt"

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

type ServicePermission interface {
	FetchServerPermissions(ctx context.Context, userID, serverID uuid.UUID) (authz.Permission, error)
}

type Service struct {
	log         *zap.Logger
	repo        Repository
	sPermission ServicePermission
	sServer     ServiceServer
}

func New(log *zap.Logger, repo Repository, sPermission ServicePermission, sServer ServiceServer) *Service {
	return &Service{
		log:         log,
		repo:        repo,
		sPermission: sPermission,
		sServer:     sServer,
	}
}

func (s *Service) CreateRole(ctx context.Context, actorID, serverID uuid.UUID, name string, permissions authz.Permission) (*models.Role, error) {
	const op = "service.roles.CreateRole"

	permissionsMask, err := s.sPermission.FetchServerPermissions(ctx, actorID, serverID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get permissions: %w", op, err)
	}
	if !authz.Has(permissionsMask, authz.PermManageRole) {
		return nil, models.ErrPermissionDenied
	}

	role, err := s.repo.CreateRole(ctx, serverID, name, permissions)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return role, nil
}

func (s *Service) GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	const op = "service.roles.GetRole"

	role, err := s.repo.GetRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return role, nil
}

func (s *Service) GetServerRoles(ctx context.Context, serverID uuid.UUID) ([]models.Role, error) {
	const op = "service.roles.GetServerRoles"

	roles, err := s.repo.GetServerRoles(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return roles, nil
}

func (s *Service) UpdateRole(ctx context.Context, actorID, serverID, roleID uuid.UUID, name *string, permission *authz.Permission) (*models.Role, error) {
	const op = "service.roles.UpdateRole"

	permissionsMask, err := s.sPermission.FetchServerPermissions(ctx, actorID, serverID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get permissions: %w", op, err)
	}
	if !authz.Has(permissionsMask, authz.PermManageRole) {
		return nil, models.ErrPermissionDenied
	}

	role, err := s.repo.UpdateRole(ctx, roleID, name, permission)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return role, nil
}

func (s *Service) ReplaceChannelPermissionOverrides(ctx context.Context, actorID, serverID, channelID uuid.UUID, overrides []models.ChannelPermissionOverride) error {
	const op = "service.roles.ReplaceChannelPermissionOverrides"

	permissionsMask, err := s.sPermission.FetchServerPermissions(ctx, actorID, serverID)
	if err != nil {
		return fmt.Errorf("%s: failed to get permissions: %w", op, err)
	}
	if !authz.Has(permissionsMask, authz.PermManageRole) {
		return models.ErrPermissionDenied
	}

	if err := s.repo.ReplaceChannelPermissionOverrides(ctx, channelID, overrides); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
