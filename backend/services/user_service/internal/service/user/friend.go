package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/model"
	"go.uber.org/zap"
)

type FriendRepository interface {
	SendFriendRequest(ctx context.Context, userID, friendID uuid.UUID) error
	FriendList(ctx context.Context, userID uuid.UUID) ([]*model.Relationship, error)
	FriendRequestList(ctx context.Context, userID uuid.UUID) ([]*model.Relationship, error)
	AcceptFriendRequest(ctx context.Context, actorID, targetID uuid.UUID) error
	DenyFriendRequest(ctx context.Context, actorID, targetID uuid.UUID) error
	BlockUser(ctx context.Context, actorID, targetID uuid.UUID) error
	RemoveFriend(ctx context.Context, actorID, targetID uuid.UUID) error
}

type UserRepository interface {
	GetUser(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type FriendEventPublisher interface {
	PublishRelationshipUpdated(ctx context.Context, userID, targetID uuid.UUID, status model.RelationshipStatus) error
	PublishRelationshipDeleted(ctx context.Context, userID, targetID uuid.UUID) error
}

type FriendService struct {
	log       *zap.Logger
	repo      FriendRepository
	userRepo  UserRepository
	publisher FriendEventPublisher
}

func NewFriendService(log *zap.Logger, repo FriendRepository, userRepo UserRepository, publisher FriendEventPublisher) *FriendService {
	return &FriendService{
		log:       log,
		repo:      repo,
		userRepo:  userRepo,
		publisher: publisher,
	}
}

func (s *FriendService) SendFriendRequest(ctx context.Context, userID, friendID uuid.UUID) error {
	const op = "service.user.SendFriendRequest"

	if userID == friendID {
		return model.ErrImpossible
	}

	err := s.repo.SendFriendRequest(ctx, userID, friendID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if s.publisher != nil {
		// Try to verify current status to publish correct event
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
			_ = s.publisher.PublishRelationshipUpdated(ctx, userID, friendID, model.StatusFriends)
			_ = s.publisher.PublishRelationshipUpdated(ctx, friendID, userID, model.StatusFriends)
		} else {
			_ = s.publisher.PublishRelationshipUpdated(ctx, userID, friendID, model.StatusRequestSent)
			_ = s.publisher.PublishRelationshipUpdated(ctx, friendID, userID, model.StatusRequestReceived)
		}
	}

	return nil
}

func (s *FriendService) FriendList(ctx context.Context, userID uuid.UUID) ([]*model.RelationshipPreview, error) {
	const op = "service.user.FriendList"

	relationships, err := s.repo.FriendList(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	previews := make([]*model.RelationshipPreview, 0, len(relationships))
	for _, r := range relationships {
		u, err := s.userRepo.GetUser(ctx, r.TargetID)
		if err != nil {
			s.log.Warn("failed to fetch target user for friend list", zap.String("op", op), zap.String("target_id", r.TargetID.String()), zap.Error(err))
			continue
		}
		previews = append(previews, &model.RelationshipPreview{
			UserID:    r.TargetID,
			Username:  u.Username,
			AvatarURL: u.AvatarURL,
			Status:    r.Status,
		})
	}

	return previews, nil
}

func (s *FriendService) FriendRequestList(ctx context.Context, userID uuid.UUID) ([]*model.RelationshipPreview, error) {
	const op = "service.user.FriendRequestList"

	relationships, err := s.repo.FriendRequestList(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	previews := make([]*model.RelationshipPreview, 0, len(relationships))
	for _, r := range relationships {
		u, err := s.userRepo.GetUser(ctx, r.TargetID)
		if err != nil {
			s.log.Warn("failed to fetch target user for request list", zap.String("op", op), zap.String("target_id", r.TargetID.String()), zap.Error(err))
			continue
		}
		previews = append(previews, &model.RelationshipPreview{
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

	if s.publisher != nil {
		_ = s.publisher.PublishRelationshipUpdated(ctx, actorID, targetID, model.StatusFriends)
		_ = s.publisher.PublishRelationshipUpdated(ctx, targetID, actorID, model.StatusFriends)
	}

	return nil
}

func (s *FriendService) DenyFriendRequest(ctx context.Context, actorID, targetID uuid.UUID) error {
	const op = "service.user.DenyFriendRequest"

	err := s.repo.DenyFriendRequest(ctx, actorID, targetID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if s.publisher != nil {
		_ = s.publisher.PublishRelationshipDeleted(ctx, actorID, targetID)
		_ = s.publisher.PublishRelationshipDeleted(ctx, targetID, actorID)
	}

	return nil
}

func (s *FriendService) BlockUser(ctx context.Context, actorID, targetID uuid.UUID) error {
	const op = "service.user.BlockUser"

	if actorID == targetID {
		return model.ErrImpossible
	}

	err := s.repo.BlockUser(ctx, actorID, targetID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if s.publisher != nil {
		_ = s.publisher.PublishRelationshipUpdated(ctx, actorID, targetID, model.StatusBlocked)
		_ = s.publisher.PublishRelationshipDeleted(ctx, targetID, actorID)
	}

	return nil
}

func (s *FriendService) RemoveFriend(ctx context.Context, actorID, targetID uuid.UUID) error {
	const op = "service.user.RemoveFriend"

	err := s.repo.RemoveFriend(ctx, actorID, targetID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if s.publisher != nil {
		_ = s.publisher.PublishRelationshipDeleted(ctx, actorID, targetID)
		_ = s.publisher.PublishRelationshipDeleted(ctx, targetID, actorID)
	}

	return nil
}
