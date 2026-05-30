package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/models"
	"go.uber.org/zap"
)

type Repository interface {
	SendFriendRequest(ctx context.Context, userID, friendID uuid.UUID) error
	FriendList(ctx context.Context, userID uuid.UUID) ([]*models.Relationship, error)
	FriendRequestList(ctx context.Context, userID uuid.UUID) ([]*models.Relationship, error)
	AcceptFriendRequest(ctx context.Context, actorID, targetID uuid.UUID) error
	DenyFriendRequest(ctx context.Context, actorID, targetID uuid.UUID) error
	BlockUser(ctx context.Context, actorID, targetID uuid.UUID) error
	RemoveFriend(ctx context.Context, actorID, targetID uuid.UUID) error
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type Broker interface {
	PublishRelationshipUpdated(ctx context.Context, userID, targetID uuid.UUID, status models.RelationshipStatus) error
	PublishRelationshipDeleted(ctx context.Context, userID, targetID uuid.UUID) error
}

type FriendService struct {
	log    *zap.Logger
	repo   Repository
	broker Broker
}

func New(log *zap.Logger, repo Repository, broker Broker) *FriendService {
	return &FriendService{
		log:    log,
		repo:   repo,
		broker: broker,
	}
}

func (s *FriendService) SendFriendRequest(ctx context.Context, userID, friendID uuid.UUID) error {
	const op = "service.user.SendFriendRequest"

	if userID == friendID {
		return models.ErrImpossible
	}

	err := s.repo.SendFriendRequest(ctx, userID, friendID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if s.broker != nil {
		relationships, err := s.repo.FriendList(ctx, userID)
		isAccepted := false
		if err == nil {
			for _, r := range relationships {
				if r.TargetID == friendID {
					isAccepted = true
					break
				}
			}
		}

		if isAccepted {
			_ = s.broker.PublishRelationshipUpdated(ctx, userID, friendID, models.StatusFriends)
			_ = s.broker.PublishRelationshipUpdated(ctx, friendID, userID, models.StatusFriends)
		} else {
			_ = s.broker.PublishRelationshipUpdated(ctx, userID, friendID, models.StatusRequestSent)
			_ = s.broker.PublishRelationshipUpdated(ctx, friendID, userID, models.StatusRequestReceived)
		}
	}

	return nil
}

func (s *FriendService) FriendList(ctx context.Context, userID uuid.UUID) ([]*models.RelationshipPreview, error) {
	const op = "service.user.FriendList"

	relationships, err := s.repo.FriendList(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	previews := make([]*models.RelationshipPreview, 0, len(relationships))
	for _, r := range relationships {
		u, err := s.repo.GetUser(ctx, r.TargetID)
		if err != nil {
			s.log.Warn("failed to fetch target user for friend list", zap.String("op", op), zap.String("target_id", r.TargetID.String()), zap.Error(err))
			continue
		}
		previews = append(previews, &models.RelationshipPreview{
			UserID:    r.TargetID,
			Username:  u.Username,
			AvatarURL: u.AvatarURL,
			Status:    r.Status,
		})
	}

	return previews, nil
}

func (s *FriendService) FriendRequestList(ctx context.Context, userID uuid.UUID) ([]*models.RelationshipPreview, error) {
	const op = "service.user.FriendRequestList"

	relationships, err := s.repo.FriendRequestList(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	previews := make([]*models.RelationshipPreview, 0, len(relationships))
	for _, r := range relationships {
		u, err := s.repo.GetUser(ctx, r.TargetID)
		if err != nil {
			s.log.Warn("failed to fetch target user for request list", zap.String("op", op), zap.String("target_id", r.TargetID.String()), zap.Error(err))
			continue
		}
		previews = append(previews, &models.RelationshipPreview{
			UserID:    r.TargetID,
			Username:  u.Username,
			AvatarURL: u.AvatarURL,
			Status:    r.Status,
		})
	}

	return previews, nil
}

func (s *FriendService) AcceptFriendRequest(ctx context.Context, actorID, targetID uuid.UUID) error {
	const op = "service.user.AcceptFriendRequest"

	err := s.repo.AcceptFriendRequest(ctx, actorID, targetID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if s.broker != nil {
		_ = s.broker.PublishRelationshipUpdated(ctx, actorID, targetID, models.StatusFriends)
		_ = s.broker.PublishRelationshipUpdated(ctx, targetID, actorID, models.StatusFriends)
	}

	return nil
}

func (s *FriendService) DenyFriendRequest(ctx context.Context, actorID, targetID uuid.UUID) error {
	const op = "service.user.DenyFriendRequest"

	err := s.repo.DenyFriendRequest(ctx, actorID, targetID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if s.broker != nil {
		_ = s.broker.PublishRelationshipDeleted(ctx, actorID, targetID)
		_ = s.broker.PublishRelationshipDeleted(ctx, targetID, actorID)
	}

	return nil
}

func (s *FriendService) BlockUser(ctx context.Context, actorID, targetID uuid.UUID) error {
	const op = "service.user.BlockUser"

	if actorID == targetID {
		return models.ErrImpossible
	}

	err := s.repo.BlockUser(ctx, actorID, targetID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if s.broker != nil {
		_ = s.broker.PublishRelationshipUpdated(ctx, actorID, targetID, models.StatusBlocked)
		_ = s.broker.PublishRelationshipDeleted(ctx, targetID, actorID)
	}

	return nil
}

func (s *FriendService) RemoveFriend(ctx context.Context, actorID, targetID uuid.UUID) error {
	const op = "service.user.RemoveFriend"

	err := s.repo.RemoveFriend(ctx, actorID, targetID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if s.broker != nil {
		_ = s.broker.PublishRelationshipDeleted(ctx, actorID, targetID)
		_ = s.broker.PublishRelationshipDeleted(ctx, targetID, actorID)
	}

	return nil
}
