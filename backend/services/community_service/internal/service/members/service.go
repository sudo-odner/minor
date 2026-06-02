package members

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	userv1 "github.com/sudo-odner/minor-shared/pkg/pb/user/v1"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
	"go.uber.org/zap"
)

type Repository interface {
	AddMember(ctx context.Context, serverID uuid.UUID, userID uuid.UUID, nickname string) (*models.Member, error)
	GetServerMember(ctx context.Context, serverID, userID uuid.UUID) (*models.Member, error)
	GetServerMembers(ctx context.Context, serverID uuid.UUID) ([]models.Member, error)
	RemoveMember(ctx context.Context, serverID, userID uuid.UUID) error
	UpdateMemberNickname(ctx context.Context, serverID, userID uuid.UUID, nickname string) error
	AddRoleToMember(ctx context.Context, serverID, userID, roleID uuid.UUID) error
	RemoveRoleFromMember(ctx context.Context, serverID, userID, roleID uuid.UUID) error
}

type UserClient interface {
    GetBatchProfiles(ctx context.Context, userIDs []string) (map[string]*userv1.UserProfile, error)
}

type PermissionService interface {
	FetchServerPermissions(ctx context.Context, userID, serverID uuid.UUID) (authz.Permission, error)
}

type ServerService interface {
	GetServer(ctx context.Context, serverID uuid.UUID) (*models.Server, error)
}

type Service struct {
	log         *zap.Logger
	repo        Repository
	userClient UserClient
	sPermission PermissionService
	sServer     ServerService
}

func New(log *zap.Logger, repo Repository, sPermission PermissionService, sServer ServerService, userClient UserClient) *Service {
	return &Service{
		log:         log,
		repo:        repo,
		sPermission: sPermission,
		sServer:     sServer,
		userClient: userClient,
	}
}

func (s *Service) AddMember(ctx context.Context, serverID, userID uuid.UUID, nickname string) (*models.Member, error) {
	const op = "service.member.AddMember"

	// Только у когдо есть модерация

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

	log := s.log.With(
		zap.String("op", op),
	)

	dbMembers, err := s.repo.GetServerMembers(ctx, serverID)
    if err != nil { return nil, err }

    fmt.Println("members:", dbMembers)
    
    // 2. Собираем все UserID в один срез
    userIDs := make([]string, len(dbMembers))
    for i, m := range dbMembers {
        userIDs[i] = m.UserID.String()
    }

    fmt.Println("IDs: ", userIDs)

    // 3. Делаем ОДИН gRPC вызов в User Service за именами и аватарами
    profiles, err := s.userClient.GetBatchProfiles(ctx, userIDs)
    if err != nil {
        log.Warn("Failed to get profiles from User Service: %v", zap.Error(err))
    }

    fmt.Println(profiles)

    // 4. Склеиваем данные (Enrichment)
    for i := range dbMembers {
        uid := dbMembers[i].UserID.String()
        if profile, ok := profiles[uid]; ok {
            dbMembers[i].Username = profile.Username   
            dbMembers[i].AvatarURL = profile.AvatarUrl 
        } else {
            // Если профиль не найден в User Service, запишем хоть что-то для отладки
            dbMembers[i].Username = "User-" + uid[:4] 
        }
    }

    return dbMembers, nil
}

func (s *Service) RemoveMember(
	ctx context.Context,
	actorID uuid.UUID,
	serverID uuid.UUID,
	targetUserID uuid.UUID,
) error {
	const op = "service.member.RemoveMember"

	server, err := s.sServer.GetServer(ctx, serverID)
	if err != nil {
		return fmt.Errorf("%s: failed to get server: %w", op, err)
	}

	if server.OwnerID == targetUserID {
		return fmt.Errorf("server owner cannnot leave or kick (wihtout transferring ownership) the server: %w", models.ErrImpossible)
	}
	if actorID != targetUserID {
		permissions, err := s.sPermission.FetchServerPermissions(ctx, actorID, serverID)
		if err != nil {
			return fmt.Errorf("%s: falied to fetch permissions: %w", op, err)
		}

		if !authz.Has(permissions, authz.PermKickMembers) {
			return models.ErrPermissionDenied
		}
	}

	if err := s.repo.RemoveMember(ctx, serverID, targetUserID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) UpdateNickname(ctx context.Context, actorID, serverID, targetUserID uuid.UUID, nickname string) error {
	const op = "service.member.UpdateNickname"

	permission, err := s.sPermission.FetchServerPermissions(ctx, actorID, serverID)
	if err != nil {
		return fmt.Errorf("%s: failed to fetch permissions: %w", op, err)
	}
	if actorID == targetUserID {
		if !authz.Has(permission, authz.PermChangeNicknames) {
			return models.ErrPermissionDenied
		}
	} else {
		if !authz.Has(permission, authz.PermManageNicknames) {
			return models.ErrPermissionDenied
		}
	}

	if err := s.repo.UpdateMemberNickname(ctx, serverID, targetUserID, nickname); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) AddRoleToMember(ctx context.Context, actorID, serverID, targetUserID, roleID uuid.UUID) error {
	const op = "service.member.AddRoleToMember"

	permissionsMask, err := s.sPermission.FetchServerPermissions(ctx, actorID, serverID)
	if err != nil {
		return fmt.Errorf("%s: failed to get permissions: %w", op, err)
	}
	if !authz.Has(permissionsMask, authz.PermManageRole) {
		return models.ErrPermissionDenied
	}

	if err := s.repo.AddRoleToMember(ctx, serverID, targetUserID, roleID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) RemoveRoleFromMember(ctx context.Context, actorID, serverID, targetUserID, roleID uuid.UUID) error {
	const op = "service.member.RemoveRoleFromMember"

	permissionsMask, err := s.sPermission.FetchServerPermissions(ctx, actorID, serverID)
	if err != nil {
		return fmt.Errorf("%s: failed to get permissions: %w", op, err)
	}
	if !authz.Has(permissionsMask, authz.PermManageRole) {
		return models.ErrPermissionDenied
	}

	if err := s.repo.RemoveRoleFromMember(ctx, serverID, targetUserID, roleID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
