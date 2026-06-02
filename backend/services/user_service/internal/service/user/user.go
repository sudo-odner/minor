package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	userv1 "github.com/sudo-odner/minor-shared/pkg/pb/user/v1"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/models"
	"go.uber.org/zap"
)

type Repository interface {
	CreateUser(ctx context.Context, u *models.User) (*models.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, username string, bio string) (*models.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetUsersByIDs(ctx context.Context, userIDs []string) ([]models.User, error)
}

type Broker interface {
	PublishUserCreated(ctx context.Context, u *models.User) error
	PublishUserUpdated(ctx context.Context, u *models.User) error
	PublishUserDeleted(ctx context.Context, userID uuid.UUID) error
}

type UserService struct {
	log    *zap.Logger
	repo   Repository
	broker Broker
}

func New(log *zap.Logger, repo Repository, broker Broker) *UserService {
	return &UserService{
		log:    log,
		repo:   repo,
		broker: broker,
	}
}

func (s *UserService) CreateUser(ctx context.Context, userID uuid.UUID, username, bio string) (*models.User, error) {
	const op = "service.user.CreateUser"

	u := &models.User{
		ID:       userID,
		Username: username,
		Bio:      bio,
	}

	created, err := s.repo.CreateUser(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if s.broker != nil {
		if err := s.broker.PublishUserCreated(ctx, created); err != nil {
			s.log.Error("failed to publish user created event", zap.String("op", op), zap.Error(err))
		}
	}

	return created, nil
}

func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	const op = "service.user.GetUser"

	u, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return u, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, username, bio string) (*models.User, error) {
	const op = "service.user.UpdateUser"

	updated, err := s.repo.UpdateUser(ctx, userID, username, bio)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if s.broker != nil {
		if err := s.broker.PublishUserUpdated(ctx, updated); err != nil {
			s.log.Error("failed to publish user updated event", zap.String("op", op), zap.Error(err))
		}
	}

	return updated, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const op = "service.user.DeleteUser"

	err := s.repo.DeleteUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if s.broker != nil {
		if err := s.broker.PublishUserDeleted(ctx, userID); err != nil {
			s.log.Error("failed to publish user deleted event", zap.String("op", op), zap.Error(err))
		}
	}

	return nil
}

func (s *UserService) GetBatchProfiles(ctx context.Context, userIDs []string) (map[string]*userv1.UserProfile, error) {
	// 1. Получаем данные из репозитория (Postgres)
	users, err := s.repo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	// 2. Конвертируем в формат Protobuf-ответа
	profiles := make(map[string]*userv1.UserProfile)
	for _, u := range users {
		profiles[u.ID.String()] = &userv1.UserProfile{
			Id:        u.ID.String(),
			Username:  u.Username,
			AvatarUrl: *u.AvatarURL,
		}
	}

	return profiles, nil
}

func (s *UserService) HandleRegistration(ctx context.Context, userID, email, username string) error {
	s.log.Info("handling registration event from nats", zap.String("user_id", userID))

	id, _ := uuid.Parse(userID)
	// Вызываем репозиторий, чтобы создать профиль в базе User Service
	_, err := s.repo.CreateUser(ctx, &models.User{ID: id, Email: email, Username: username})
	if err != nil {
		return fmt.Errorf("service.user.HandleRegistration: %w", err)
	}

	return nil
}