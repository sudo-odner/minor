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
	UpdateChannel(ctx context.Context, channelID, serverID uuid.UUID, name *string, parentID *uuid.UUID) (*models.Channel, error)
	DeleteChannel(ctx context.Context, channelID uuid.UUID) error
	MoveChannel(ctx context.Context, serverID, channelID uuid.UUID, oldParentID, newParentID *uuid.UUID, oldPos, newPos int) error
}

type Broker interface {
	PublishChannelCreated(ctx context.Context, serverID uuid.UUID, ch models.Channel) error
	PublishChannelUpdated(ctx context.Context, serverID uuid.UUID, ch models.Channel) error
	PublishChannelDeleted(ctx context.Context, serverID, channelID uuid.UUID) error
	PublishChannelPositionsUpdated(ctx context.Context, serverID uuid.UUID, channels []models.Channel) error
}

type PermissionService interface {
	FetchServerPermission(ctx context.Context, userID, serverID uuid.UUID) (authz.Permission, error)
}

type Service struct {
	log    *zap.Logger
	repo   Repository
	broker Broker
	perm   PermissionService
}

func New(log *zap.Logger, repo Repository, broker Broker, permService PermissionService) *Service {
	return &Service{
		log:    log,
		repo:   repo,
		broker: broker,
		perm:   permService,
	}
}

func (s *Service) CreateChannel(
	ctx context.Context,
	actorID uuid.UUID,
	serverID uuid.UUID,
	name string,
	typeChannel models.ChannelType,
	parentID *uuid.UUID,
) (*models.Channel, error) {
	const op = "service.channel.CreateChannel"

	permission, err := s.perm.FetchServerPermission(ctx, actorID, serverID)
	if err != nil {
		return nil, fmt.Errorf("%s: falied to fetch permissions: %w", op, err)
	}
	if !authz.Has(permission, authz.PermManageChannels) {
		return nil, models.ErrPermissionDenied
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

	_ = s.broker.PublishChannelCreated(ctx, serverID, *ch)

	return ch, nil
}

func (s *Service) GetChannel(ctx context.Context, channelID uuid.UUID) (*models.Channel, error) {
	return nil, nil
}

func (s *Service) GetServerChannel(ctx context.Context, serverID uuid.UUID) (*models.Channel, error) {
	return nil, nil
}

func (s *Service) UpdatedChannel(
	ctx context.Context,
	serverID uuid.UUID,
	name string,
	typeChannel models.ChannelType,
	parentID *uuid.UUID,
) (*models.Channel, error) {
	return nil, nil
}

func (s *Service) DeleteChannel(ctx context.Context, serverID, channelID uuid.UUID) error {
	return nil
}

func (s *Service) MoveChannel(ctx context.Context, serverID, channelID uuid.UUID, oldParentID, newParentID *uuid.UUID, oldPos, newPos int) error {
	return nil
}
