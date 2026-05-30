package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/model"
	"go.uber.org/zap"
)

type Repository interface {
	CreateUser(ctx context.Context, u *model.User) (*model.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*model.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, username *string, bio *string) (*model.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type EventPublisher interface {
	PublishUserCreated(ctx context.Context, u *model.User) error
	PublishUserUpdated(ctx context.Context, u *model.User) error
	PublishUserDeleted(ctx context.Context, userID uuid.UUID) error
}

type Service struct {
	log       *zap.Logger
	repo      Repository
	publisher EventPublisher
}

func New(log *zap.Logger, repo Repository, publisher EventPublisher) *Service {
	return &Service{
		log:       log,
		repo:      repo,
		publisher: publisher,
	}
}

func (s *Service) CreateUser(ctx context.Context, userID uuid.UUID, username, bio string) (*model.User, error) {
	const op = "service.user.CreateUser"

	u := &model.User{
		ID:       userID,
		Username: username,
		Bio:      bio,
	}

	created, err := s.repo.CreateUser(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if s.publisher != nil {
		if err := s.publisher.PublishUserCreated(ctx, created); err != nil {
			s.log.Error("failed to publish user created event", zap.String("op", op), zap.Error(err))
		}
	}

	return created, nil
}

func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	const op = "service.user.GetUser"

	u, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return u, nil
}

func (s *Service) UpdateUser(ctx context.Context, userID uuid.UUID, username, bio *string) (*model.User, error) {
	const op = "service.user.UpdateUser"

	updated, err := s.repo.UpdateUser(ctx, userID, username, bio)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if s.publisher != nil {
		if err := s.publisher.PublishUserUpdated(ctx, updated); err != nil {
			s.log.Error("failed to publish user updated event", zap.String("op", op), zap.Error(err))
		}
	}

	return updated, nil
}

func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const op = "service.user.DeleteUser"

	err := s.repo.DeleteUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if s.publisher != nil {
		if err := s.publisher.PublishUserDeleted(ctx, userID); err != nil {
			s.log.Error("failed to publish user deleted event", zap.String("op", op), zap.Error(err))
		}
	}

	return nil
}
