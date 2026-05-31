package presence

import (
	"context"
	"fmt"
	"time"

	"github.com/sudo-odner/minor/backend/services/presence_service/internal/models"
	"go.uber.org/zap"
)

type Cache interface {
	SetStatus(ctx context.Context, userID string, status models.UserStatus, customStatus string, lastActiveAt int64) error
	GetUserStatuses(ctx context.Context, userIDs []string) (map[string]*models.Presence, error)
}

type Broker interface {
	PublishPresenceStatusUpdated(ctx context.Context, p *models.Presence) error
}

type Service struct {
	log    *zap.Logger
	cache  Cache
	broker Broker
}

func New(log *zap.Logger, cache Cache, broker Broker) *Service {
	return &Service{
		log:    log,
		cache:  cache,
		broker: broker,
	}
}

func (s *Service) SetStatus(ctx context.Context, userID string, status models.UserStatus, customStatus string) error {
	const op = "service.presence.SetStatus"
	s.log.Debug("setting status", zap.String("user_id", userID), zap.Any("status", status))

	lastActiveAt := time.Now().Unix()

	err := s.cache.SetStatus(ctx, userID, status, customStatus, lastActiveAt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	p := &models.Presence{
		UserID:       userID,
		Status:       status,
		CustomStatus: customStatus,
		LastActiveAt: lastActiveAt,
	}

	if s.broker != nil {
		err = s.broker.PublishPresenceStatusUpdated(ctx, p)
		if err != nil {
			s.log.Warn("failed to publish presence status updated event", zap.Error(err))
		}
	}

	return nil
}

func (s *Service) GetUserStatuses(ctx context.Context, userIDs []string) (map[string]*models.Presence, error) {
	const op = "service.presence.GetUserStatuses"
	s.log.Debug("getting user statuses", zap.Int("count", len(userIDs)))

	statuses, err := s.cache.GetUserStatuses(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return statuses, nil
}
