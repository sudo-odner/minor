package servers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
	"go.uber.org/zap"
)

type Repository interface {
	CreateServerWithDefaultSetup(ctx context.Context, name string, ownerID uuid.UUID, avatarURL string) (*models.Server, error)
	GetServer(ctx context.Context, serverID uuid.UUID) (*models.Server, error)
	UpdateServer(ctx context.Context, serverID uuid.UUID, name *string, avatarURL *string) (*models.Server, error)
	DeleteServer(ctx context.Context, serverID uuid.UUID) error
	GetUserServers(ctx context.Context, userID uuid.UUID) ([]models.Server, error)
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

// AvatarURL не работает на данный момент всегда идет записть "" так как нет сервса который хранит картинки
// TODO: Написать Events под манипуляциями над сервером, чтобы обновлялось у всех

func (s *Service) CreateServer(ctx context.Context, name string, ownerID uuid.UUID, avatarURL string) (*models.Server, error) {
	const op = "service.server.CreateServer"

	server, err := s.repo.CreateServerWithDefaultSetup(ctx, name, ownerID, "")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return server, nil
}

func (s *Service) GetServer(ctx context.Context, serverID uuid.UUID) (*models.Server, error) {
	const op = "service.server.GetServer"

	server, err := s.repo.GetServer(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return server, nil
}

func (s *Service) UpdateServer(ctx context.Context, actorID uuid.UUID, serverID uuid.UUID, name *string, avatarURL *string) (*models.Server, error) {
	const op = "service.server.UpdateServer"

	server, err := s.repo.GetServer(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if server.OwnerID != actorID {
		return nil, models.ErrPermissionDenied
	}

	updated, err := s.repo.UpdateServer(ctx, serverID, name, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return updated, nil
}

func (s *Service) DeleteServer(ctx context.Context, actorID uuid.UUID, serverID uuid.UUID) error {
	const op = "service.server.DeleteServer"

	server, err := s.repo.GetServer(ctx, serverID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if server.OwnerID != actorID {
		return models.ErrPermissionDenied
	}

	if err := s.repo.DeleteServer(ctx, serverID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) GetUserServers(ctx context.Context, userID uuid.UUID) ([]models.Server, error) {
	const op = "service.server.GetUserServers"

	serversList, err := s.repo.GetUserServers(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return serversList, nil
}
