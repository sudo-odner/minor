package channel

import (
	"context"
	"fmt"
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
	Members(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error)
	UserPermission(ctx context.Context, channelID, userID uuid.UUID) (authz.Permission, error)
	Delete(ctx context.Context, channelID, userID uuid.UUID) error
}

type EventPublisher interface {
	Publish(ctx context.Context, evt events.Event)
}

type UserFetcher interface {
	NameByID(ctx context.Context, userID uuid.UUID) (string, error)
}

type ChannelService struct {
	log               *slog.Logger
	channelRepository ChannelRepository
	eventPublisher    EventPublisher
	userFetcher       UserFetcher
}

func NewChannelService(log *slog.Logger, channelRepository ChannelRepository, eventPublisher EventPublisher, userFetcher UserFetcher) *ChannelService {
	return &ChannelService{
		log:               log,
		channelRepository: channelRepository,
		eventPublisher:    eventPublisher,
		userFetcher:       userFetcher,
	}
}

func (s *ChannelService) UserPermission(ctx context.Context, channelID, userID uuid.UUID) (authz.Permission, error) {
	const op = "service.channel.UserPermission"
	permission, err := s.channelRepository.UserPermission(ctx, channelID, userID)
	if err != nil {
		return permission, fmt.Errorf("%s: %w", op, err)
	}
	return permission, nil
}

func (s *ChannelService) Members(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
	const op = "service.channel.Members"
	members, err := s.channelRepository.Members(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return members, nil
}

func (s *ChannelService) CreateDM(ctx context.Context, actorID, partnerID uuid.UUID) (*domain.Channel, error) {
	const op = "serivce.channel.CreateDM"

	name, err := s.userFetcher.NameByID(ctx, actorID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	channel, err := s.channelRepository.Create(ctx, domain.ChannelTypeDM, nil, []uuid.UUID{actorID, partnerID})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	channel.Name = &name
	return channel, nil
}

func (s *ChannelService) CreateDMGroup(ctx context.Context, actorID uuid.UUID, name string, userIDs []uuid.UUID) (*domain.Channel, error) {
	const op = "service.channel.CreateDMGroup"

	channel, err := s.channelRepository.Create(ctx, domain.ChannelTypeDMGroup, &name, userIDs)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return channel, nil
}
