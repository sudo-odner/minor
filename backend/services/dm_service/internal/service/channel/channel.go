package channel

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	"github.com/sudo-odner/minor-shared/pkg/nats/events"
	"github.com/sudo-odner/minor/backend/service/dm_service/internal/domain"
)

type ChannelRepository interface {
	Create(ctx context.Context, channelType domain.ChannelType, name *string, userIDs []uuid.UUID) (*domain.Channel, error)
	ByID(ctx context.Context, channelID uuid.UUID) (*domain.Channel, error)
	ByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Channel, error)
	Membres(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error)
	UserPermission(ctx context.Context, channelID, userID uuid.UUID) (authz.Permission, error)
	Delete(ctx context.Context, channelID, userID uuid.UUID) error
}

type EventPublisher interface {
	Publish(ctx context.Context, evt events.Event)
}

type ChannelService struct {
	log               *slog.Logger
	channelRepository ChannelRepository
	eventPublisher    EventPublisher
}

func New(log *slog.Logger, channelRepository ChannelRepository, eventPublisher EventPublisher) *ChannelService {
	return &ChannelService{
		log:               log,
		channelRepository: channelRepository,
		eventPublisher:    eventPublisher,
	}
}
