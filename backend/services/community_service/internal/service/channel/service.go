package channel

import (
	"context"

	"github.com/google/uuid"
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
type Service struct {
	log    *zap.Logger
	repo   Repository
	broker Broker
}

func New(log *zap.Logger, repo Repository, broker Broker) *Service {
	return &Service{
		log:    log,
		repo:   repo,
		broker: broker,
	}
}

func (s *Service) CreateChannel(
	ctx context.Context,
	serverID uuid.UUID,
	name string,
	typeChannel models.ChannelType,
	parentID *uuid.UUID,
) (*models.Channel, error) {
	return nil, nil
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
