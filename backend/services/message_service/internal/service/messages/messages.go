package messages

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/models"
	"go.uber.org/zap"
)

type Repository interface {
	SaveMessage(ctx context.Context, authorID, channelID uuid.UUID, content string, replyTo *uuid.UUID) (*models.Message, error)
	GetMessages(ctx context.Context, channelID uuid.UUID, limit int, beforeID *uuid.UUID) ([]models.Message, error)
	GetMessage(ctx context.Context, channelID, messageID uuid.UUID) (*models.Message, error)
	DeleteMessage(ctx context.Context, channelID uuid.UUID, messageID uuid.UUID) error
}

type Cache interface {
	GetChannelOwner(ctx context.Context, channelID uuid.UUID) (models.ChannelOwner, error)
	WriteChannelOwner(ctx context.Context, channelID uuid.UUID, owner models.ChannelOwner) error
}

type BrokerProducer interface {
	PublishMessageCreated(ctx context.Context, msg *models.Message) error
	PublishMessageDeleted(ctx context.Context, channelID, messageID uuid.UUID) error
}

type CommunityClient interface {
	FetchPermission(ctx context.Context, userID, channelID uuid.UUID) (authz.Permission, error)
	CheckChannelExists(ctx context.Context, channelID uuid.UUID) (bool, error)
}

type DMClient interface {
	FetchPermission(ctx context.Context, userID, channelID uuid.UUID) (authz.Permission, error)
	CheckChannelExists(ctx context.Context, channelID uuid.UUID) (bool, error)
}

type MessageService struct {
	log             *zap.Logger
	repo            Repository
	brokerProducer  BrokerProducer
	cache           Cache
	communityClient CommunityClient
	dmClient        DMClient
}

func New(log *zap.Logger,
	repo Repository,
	brokerProducer BrokerProducer,
	cache Cache,
	communityClient CommunityClient, dmClient DMClient,
) *MessageService {
	return &MessageService{
		log:             log,
		repo:            repo,
		brokerProducer:  brokerProducer,
		cache:           cache,
		communityClient: communityClient,
		dmClient:        dmClient,
	}
}

// loadChannelOwner load channel owner.
//
// Fast path: get channel owner from cache. If cache hit return channel owner
//
// Slow path: If cache miss, try ask community and user service, then write in cache metadata.
func (ms *MessageService) loadChannelOwner(
	ctx context.Context,
	channelID uuid.UUID,
) (models.ChannelOwner, error) {
	const op = "service.messages.loadChannelOwner"

	// Fast path: fetch metadata from cache
	channelOwner, err := ms.cache.GetChannelOwner(ctx, channelID)
	if err == nil {
		return channelOwner, nil
	}

	// Slow path: query Community service
	cCtx, cCancel := context.WithTimeout(ctx, 2*time.Second)
	defer cCancel()

	ok, err := ms.communityClient.CheckChannelExists(cCtx, channelID)
	if err != nil {
		return "", fmt.Errorf("%s (community grpc): %w", op, err)
	}
	if ok {
		_ = ms.cache.WriteChannelOwner(ctx, channelID, models.ChannelOwnerCommunity)
		return models.ChannelOwnerCommunity, nil
	}

	// Query DM service
	dmCtx, dmCancel := context.WithTimeout(ctx, 2*time.Second)
	defer dmCancel()

	ok, err = ms.dmClient.CheckChannelExists(dmCtx, channelID)
	if err != nil {
		return "", fmt.Errorf("%s (dm grpc): %w", op, err)
	}
	if ok {
		_ = ms.cache.WriteChannelOwner(ctx, channelID, models.ChannelOwnerDM)
		return models.ChannelOwnerDM, nil
	}

	return "", fmt.Errorf("%s: channel owner not found: %w", op, models.ErrChannelNotFound)
}

// loadPermissionMask Get an access mask. Encapsulates the logic for routing through third-party services
func (ms *MessageService) loadPermissionMask(ctx context.Context, userID, channelID uuid.UUID) (authz.Permission, error) {
	const op = "service.messages.loadPermissionMask"

	channelOwner, err := ms.loadChannelOwner(ctx, channelID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var maskPermission authz.Permission
	switch channelOwner {
	case models.ChannelOwnerCommunity:
		mask, err := ms.communityClient.FetchPermission(ctx, userID, channelID)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
		maskPermission = mask
	case models.ChannelOwnerDM:
		mask, err := ms.dmClient.FetchPermission(ctx, userID, channelID)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
		maskPermission = mask
	}
	return maskPermission, nil
}

// SaveMessage save message in cassandra and publish event in nats.
func (ms *MessageService) SaveMessage(
	ctx context.Context,
	userID, channelID uuid.UUID,
	content string,
	replyTo *uuid.UUID,
) (*models.Message, error) {
	const op = "service.messages.SaveMessage"
	log := ms.log.With(zap.String("op", op))

	maskPermission, err := ms.loadPermissionMask(ctx, userID, channelID)
	if err != nil {
		if errors.Is(err, models.ErrChannelNotFound) {
			log.Debug("channel owner not found", zap.String("channel_id", channelID.String()))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		log.Error("failed to resolve channel permissions", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	// Check access to write in channel
	if !authz.Has(maskPermission, authz.PermSendMessages) {
		log.Debug(
			"permission denied: user cannot send messages to channel",
			zap.String("user_id", userID.String()),
			zap.String("channel_id", channelID.String()),
		)
		return nil, models.ErrPermissionDenied
	}

	msg, err := ms.repo.SaveMessage(ctx, userID, channelID, content, replyTo)
	if err != nil {
		log.Error("failed save message", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := ms.brokerProducer.PublishMessageCreated(ctx, msg); err != nil {
		log.Error("failed to publish message created event to broker", zap.Error(err))
	}

	return msg, nil
}

// GetMessages retrieves up to limit messages from the specified channel in reverse chronological order.
// If beforeID is provided, it returns messages created prior to that message ID for pagination.
func (ms *MessageService) GetMessages(ctx context.Context, userID uuid.UUID, channelID uuid.UUID, limit int, beforeID *uuid.UUID) ([]models.Message, error) {
	const op = "service.messages.GetMessages"
	log := ms.log.With(
		zap.String("op", op),
	)

	maskPermission, err := ms.loadPermissionMask(ctx, userID, channelID)
	if err != nil {
		if errors.Is(err, models.ErrChannelNotFound) {
			log.Debug("channel owner not found", zap.String("channel_id", channelID.String()))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		log.Error("failed to resolve channel permissions", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	// Check access to read message
	if !authz.Has(maskPermission, authz.PermViewChannel) {
		log.Debug(
			"permission denied: user cannot read messages in channel",
			zap.String("user_id", userID.String()),
			zap.String("channel_id", channelID.String()),
		)
		return nil, models.ErrPermissionDenied
	}

	msgs, err := ms.repo.GetMessages(ctx, channelID, limit, beforeID)
	if err != nil {
		log.Error("failed to get messages from database", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return msgs, nil
}

// GetMessage retrieves a single message from the specified channel by messageID.
func (ms *MessageService) GetMessage(
	ctx context.Context,
	userID, channelID, messageID uuid.UUID,
) (*models.Message, error) {
	const op = "service.messages.GetMessage"
	log := ms.log.With(zap.String("op", op))

	maskPermission, err := ms.loadPermissionMask(ctx, userID, channelID)
	if err != nil {
		if errors.Is(err, models.ErrChannelNotFound) {
			log.Debug("channel owner not found", zap.String("channel_id", channelID.String()))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		log.Error("failed to resolve channel permissions", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if !authz.Has(maskPermission, authz.PermViewChannel) {
		log.Debug(
			"permission denied: user cannot read messages in channel",
			zap.String("user_id", userID.String()),
			zap.String("channel_id", channelID.String()),
		)
		return nil, models.ErrPermissionDenied
	}

	msg, err := ms.repo.GetMessage(ctx, channelID, messageID)
	if err != nil {
		if errors.Is(err, models.ErrMessageNotFound) {
			log.Debug("message not found", zap.String("channel_id", channelID.String()))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		log.Error("failed to get message from database", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return msg, nil
}

// DeleteMessage deletes a message from a channel after validating permissions.
// The user must be able to view the channel and either be the message author
// or possess message management / administrator privileges.
func (ms *MessageService) DeleteMessage(ctx context.Context, userID, channelID, messageID uuid.UUID) error {
	const op = "service.messages.DeleteMessage"
	log := ms.log.With(zap.String("op", op))

	maskPermission, err := ms.loadPermissionMask(ctx, userID, channelID)
	if err != nil {
		if errors.Is(err, models.ErrChannelNotFound) {
			log.Debug("channel owner not found", zap.String("channel_id", channelID.String()))
			return fmt.Errorf("%s: %w", op, err)
		}
		log.Error("failed to resolve channel permissions", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	if !authz.Has(maskPermission, authz.PermViewChannel) {
		log.Debug(
			"permission denied: user cannot view channel",
			zap.String("user_id", userID.String()),
			zap.String("channel_id", channelID.String()),
		)
		return models.ErrPermissionDenied
	}

	msg, err := ms.repo.GetMessage(ctx, channelID, messageID)
	if err != nil {
		if errors.Is(err, models.ErrMessageNotFound) {
			log.Debug("not found message", zap.String("message_id", messageID.String()))
			return fmt.Errorf("%s: %w", op, err)
		}
		log.Error("failed get message from database", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	isAuthor := msg.UserID == userID
	canManage := authz.Has(maskPermission, authz.PermMenageMessage)

	if !isAuthor && !canManage {
		log.Debug(
			"permission denied: user is not author and lacks manage permissions",
			zap.String("user_id", userID.String()),
			zap.String("channel_id", channelID.String()),
			zap.String("message_id", messageID.String()),
		)
		return models.ErrPermissionDenied
	}

	if err = ms.repo.DeleteMessage(ctx, channelID, messageID); err != nil {
		if errors.Is(err, models.ErrMessageNotFound) {
			log.Debug("message already deleted or not found", zap.String("message_id", messageID.String()))
			return fmt.Errorf("%s: %w", op, err)
		}
		log.Error("failed to delete message from repository", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = ms.brokerProducer.PublishMessageDeleted(ctx, channelID, messageID); err != nil {
		log.Error("failed to publish delete to broker", zap.Error(err))
	}

	return nil
}

// TODO: Implement async batcher save and delete message
