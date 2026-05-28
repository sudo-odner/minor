package permissions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
	"go.uber.org/zap"
)

type ServiceServer interface {
	GetServer(ctx context.Context, serverID uuid.UUID) (*models.Server, error)
}

type ChannelRepository interface {
	GetChannel(ctx context.Context, channelID uuid.UUID) (*models.Channel, error)
}

type Repository interface {
	GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error)
	GetMemberRoles(ctx context.Context, serverID, userID uuid.UUID) ([]models.Role, error)
	GetChannelOverrides(ctx context.Context, channelID uuid.UUID) ([]models.ChannelPermissionOverride, error)
}

type Service struct {
	log      *zap.Logger
	repo     Repository
	sService ServiceServer
}

func New(log *zap.Logger, repo Repository, sService ServiceServer) *Service {
	return &Service{
		log:      log,
		repo:     repo,
		sService: sService,
	}
}

// FetchServerPermissions вычисляет базовые права пользователя на сервере (без учета каналов)
func (s *Service) FetchServerPermissions(ctx context.Context, userID, serverID uuid.UUID) (authz.Permission, error) {
	const op = "service.permissions.FetchServerPermissions"

	// 1. Проверяем владельца сервера
	server, err := s.sService.GetServer(ctx, serverID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	if server.OwnerID == userID {
		return 0xFFFFFFFFFFFFFFFF, nil
	}

	// 2. Загружаем роль @everyone (ее ID равен ID сервера)
	everyoneRole, err := s.repo.GetRole(ctx, serverID)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get @everyone role: %w", op, err)
	}

	// 3. Загружаем остальные роли пользователя
	memberRoles, err := s.repo.GetMemberRoles(ctx, serverID, userID)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get member roles: %w", op, err)
	}

	// 4. Складываем базовые права
	permissions := everyoneRole.Permission
	for _, role := range memberRoles {
		permissions |= role.Permission
	}

	return permissions, nil
}
