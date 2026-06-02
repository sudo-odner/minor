package channel

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
	"go.uber.org/zap"
)

type Repository interface {
	CreateChannel(ctx context.Context, serverID uuid.UUID, name string, typeChannel models.ChannelType, parentID *uuid.UUID) (*models.Channel, error)
	GetChannel(ctx context.Context, channelID uuid.UUID) (*models.Channel, error)
	GetServerChannels(ctx context.Context, serverID uuid.UUID) ([]models.Channel, error)
	UpdateChannel(ctx context.Context, channelID, serverID uuid.UUID, name string, parentID *uuid.UUID) (*models.Channel, error)
	DeleteChannel(ctx context.Context, channelID uuid.UUID) error
	MoveChannel(ctx context.Context, serverID, channelID uuid.UUID, oldParentID, newParentID *uuid.UUID, oldPos, newPos int) error
}

type Broker interface {
	PublishChannelCreated(ctx context.Context, serverID uuid.UUID, ch *models.Channel) error
	PublishChannelUpdated(ctx context.Context, serverID uuid.UUID, ch *models.Channel) error
	PublishChannelDeleted(ctx context.Context, serverID, channelID uuid.UUID) error
	PublishChannelPositionsUpdated(ctx context.Context, serverID uuid.UUID, channels []models.Channel) error
}

type PermissionService interface {
	FetchServerPermissions(ctx context.Context, userID, serverID uuid.UUID) (authz.Permission, error)
}

type Service struct {
	log         *zap.Logger
	repo        Repository
	broker      Broker
	sPermission PermissionService
}

func New(log *zap.Logger, repo Repository, broker Broker, sPermission PermissionService) *Service {
	return &Service{
		log:         log,
		repo:        repo,
		broker:      broker,
		sPermission: sPermission,
	}
}

// Пока проверка управлением каналами глобально(не для конкретного сервера)

func (s *Service) CreateChannel(
	ctx context.Context,
	actorID uuid.UUID,
	serverID uuid.UUID,
	name string,
	typeChannel models.ChannelType,
	parentID *uuid.UUID,
) (*models.Channel, error) {
	const op = "service.channel.CreateChannel"

	permission, err := s.sPermission.FetchServerPermissions(ctx, actorID, serverID)
	if err != nil {
		return nil, fmt.Errorf("%s: falied to fetch permissions: %w", op, err)
	}
	if !authz.Has(permission, authz.PermManageChannels) {
		return nil, models.ErrPermissionDenied
	}

	if typeChannel == models.ChannelTypeCategory && parentID != nil && *parentID != uuid.Nil {
		return nil, fmt.Errorf("category channels cannot have a parent category: %w", models.ErrImpossible)
	}

	if parentID != nil && *parentID != uuid.Nil {
		parent, err := s.repo.GetChannel(ctx, *parentID)
		if err != nil {
			return nil, fmt.Errorf("%s: parent category validation: %w", op, err)
		}
		if parent.ServerID != serverID {
			return nil, fmt.Errorf("parent category must belong to the same server: %w", models.ErrImpossible)
		}
		if parent.Type != models.ChannelTypeCategory {
			return nil, fmt.Errorf("parent channel must be of type category: %w", models.ErrImpossible)
		}
	}

	ch, err := s.repo.CreateChannel(ctx, serverID, name, typeChannel, parentID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_ = s.broker.PublishChannelCreated(ctx, serverID, ch)

	return ch, nil
}

// TODO: полумать на счет скрытых каналов

func (s *Service) GetChannel(ctx context.Context, channelID uuid.UUID) (*models.Channel, error) {
	const op = "service.channel.GetChannel"

	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return ch, nil
}

func (s *Service) GetServerChannel(ctx context.Context, serverID uuid.UUID) ([]models.Channel, error) {
	const op = "service.channel.GetServerChannel"

	chs, err := s.repo.GetServerChannels(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return chs, nil
}

func (s *Service) UpdateChannel(
	ctx context.Context,
	actorID uuid.UUID,
	channelID, serverID uuid.UUID,
	name string,
	parentID *uuid.UUID,
) (*models.Channel, error) {
	const op = "service.channel.UpdateChannel"

	permission, err := s.sPermission.FetchServerPermissions(ctx, actorID, serverID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to fetch permission: %w", op, err)
	}
	if !authz.Has(permission, authz.PermManageChannels) {
		return nil, models.ErrPermissionDenied
	}

	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if ch.ServerID != serverID {
		return nil, models.ErrPermissionDenied
	}

	if ch.Type == models.ChannelTypeCategory && parentID != nil && *parentID != uuid.Nil {
		return nil, fmt.Errorf("category channels cannot have a parent category: %w", models.ErrImpossible)
	}

	if parentID != nil {
		if *parentID == channelID {
			return nil, fmt.Errorf("channel cannot be its own parent: %w", models.ErrImpossible)
		}
		if *parentID != uuid.Nil {
			parent, err := s.repo.GetChannel(ctx, *parentID)
			if err != nil {
				return nil, fmt.Errorf("%s: parent category not found: %w", op, err)
			}
			if parent.ServerID != serverID {
				return nil, fmt.Errorf("parent category must belogn to the same server: %w", models.ErrImpossible)
			}
			if parent.Type != models.ChannelTypeCategory {
				return nil, fmt.Errorf("parent channel must be a category: %w", models.ErrImpossible)
			}
		}
	}

	updated, err := s.repo.UpdateChannel(ctx, channelID, serverID, name, parentID)
	if err != nil {
		return nil, err
	}

	_ = s.broker.PublishChannelUpdated(ctx, serverID, updated)

	return updated, nil
}

func (s *Service) DeleteChannel(ctx context.Context, actorID, serverID, channelID uuid.UUID) error {
	const op = "service.channel.DeleteChannel"

	permission, err := s.sPermission.FetchServerPermissions(ctx, actorID, serverID)
	if err != nil {
		return fmt.Errorf("%s: falied to fetch permission: %w", op, err)
	}
	if !authz.Has(permission, authz.PermManageChannels) {
		return models.ErrPermissionDenied
	}

	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return err
	}
	if ch.ServerID != serverID {
		return models.ErrPermissionDenied
	}
	if err := s.repo.DeleteChannel(ctx, channelID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_ = s.broker.PublishChannelDeleted(ctx, serverID, channelID)

	return nil
}

func (s *Service) MoveChannel(
	ctx context.Context,
	actorID uuid.UUID,
	serverID, channelID uuid.UUID,
	newParentID *uuid.UUID,
	newPos int,
) error {
	const op = "server.channel.MoveChannel"

	permission, err := s.sPermission.FetchServerPermissions(ctx, actorID, serverID)
	if err != nil {
		return fmt.Errorf("%s: falied to fetch permissions: %w", op, err)
	}
	if !authz.Has(permission, authz.PermManageChannels) {
		return models.ErrPermissionDenied
	}

	// Fetch actual current state of the channel to get true oldParentID and oldPos
	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return fmt.Errorf("%s: failed to get channel: %w", op, err)
	}
	if ch.ServerID != serverID {
		return fmt.Errorf("channel does not belong to server: %w", models.ErrImpossible)
	}

	// Retrieve actual oldParentID and oldPos from DB to prevent malicious or out-of-sync moves
	oldParentID := ch.ParentID
	oldPos := ch.Position

	if ch.Type == models.ChannelTypeCategory && newParentID != nil && *newParentID != uuid.Nil {
		return fmt.Errorf("category channels cannot have a parent category: %w", models.ErrImpossible)
	}

	if newParentID != nil {
		if *newParentID == channelID {
			return fmt.Errorf("channel cannot be its own parent: %w", models.ErrImpossible)
		}
		if *newParentID != uuid.Nil {
			parent, err := s.repo.GetChannel(ctx, *newParentID)
			if err != nil {
				return fmt.Errorf("%s: new parent category not found: %w", op, err)
			}
			if parent.ServerID != serverID {
				return fmt.Errorf("new parent category must belong to the same server: %w", models.ErrImpossible)
			}
			if parent.Type != models.ChannelTypeCategory {
				return fmt.Errorf("new parent channel must be a category: %w", models.ErrImpossible)
			}
		}
	}

	if err := s.repo.MoveChannel(ctx, serverID, channelID, oldParentID, newParentID, oldPos, newPos); err != nil {
		return fmt.Errorf("%s: db move failed: %w", op, err)
	}

	channels, err := s.repo.GetServerChannels(ctx, serverID)
	if err != nil {
		return fmt.Errorf("%s: falied to fetch updated channels: %w", op, err)
	}

	_ = s.broker.PublishChannelPositionsUpdated(ctx, serverID, channels)

	return nil
}
