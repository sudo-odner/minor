package permissions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
	"go.uber.org/zap"
)

type Repository interface {
	GetServer(ctx context.Context, serverID uuid.UUID) (*models.Server, error)
	GetChannel(ctx context.Context, channelID uuid.UUID) (*models.Channel, error)
	GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error)
	GetMemberRoles(ctx context.Context, serverID, userID uuid.UUID) ([]models.Role, error)
	GetChannelOverrides(ctx context.Context, channelID uuid.UUID) ([]models.ChannelPermissionOverride, error)
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

// FetchServerPermissions вычисляет базовые права пользователя на сервере (без учета каналов)
func (s *Service) FetchServerPermissions(ctx context.Context, userID, serverID uuid.UUID) (authz.Permission, error) {
	const op = "service.permissions.FetchServerPermissions"

	// 1. Проверяем владельца сервера
	server, err := s.repo.GetServer(ctx, serverID)
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

// Базовые настройки сервера (права на уровне сервера) -> Переопределение для базовой роли @everyone ->
// -> Применить переопределение для всех ролей пользователя -> накопить и переопределить для пользоватея

// FetchPermissions вычесляет права пользователя для канала
func (s *Service) FetchPermissions(ctx context.Context, userID, channelID uuid.UUID) (authz.Permission, error) {
	const op = "service.permissions.FetchPermissions"

	// 1. Оперделяем какому серверу пренадлежит channelID
	channel, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return 0, fmt.Errorf("%s: falied to get channel: %w", op, err)
	}
	serverID := channel.ServerID

	// 2. Получаем глобальный права пользователя
	permissions, err := s.FetchServerPermissions(ctx, userID, channelID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Интересная лгика но пока хз, нужно мнения поспрашивать так как в теории может будет такая
	// роль типо супер ролькоторая будет так же с максимальными правами
	// // Если это owner у него и так все правa
	// if permissions == 0xFFFFFFFFFFFFFFFF {
	// 	return permissions, nil
	// }

	// 3. Накидываем переопределение роли @everyone
	overwrites, err := s.repo.GetChannelOverrides(ctx, channelID)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get overrides: %w", op, err)
	}
	for _, ov := range overwrites {
		if ov.TargetType == models.OverrideTypeRole && ov.TargetID == serverID {
			permissions = (permissions & ^ov.Deny) | ov.Allow
			break
		}
	}

	// 4. Накидываем переопределение для всех рольей пользователя
	memberRoles, err := s.repo.GetMemberRoles(ctx, serverID, userID)
	if err != nil {
		return 0, err
	}

	var rolesAllow, rolesDeny authz.Permission
	for _, ov := range overwrites {
		if ov.TargetType == models.OverrideTypeRole && ov.TargetID != serverID {
			for _, role := range memberRoles {
				if role.ID == ov.TargetID {
					rolesAllow |= ov.Allow
					rolesDeny |= ov.Deny
				}
			}
		}
	}
	permissions = (permissions & ^rolesDeny) | rolesAllow

	// 5. Находим ограничения для пользователя и накидываем на него
	for _, ov := range overwrites {
		if ov.TargetType == models.OverrideTypeUser && ov.TargetID == userID {
			permissions = (permissions & ^ov.Deny) | ov.Allow
			break
		}
	}

	return permissions, nil
}
