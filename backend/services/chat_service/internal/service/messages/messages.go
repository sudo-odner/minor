package messages

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/chat_service/internal/models"
	"go.uber.org/zap"
)

type MessageRepo interface {
	SaveMessage(ctx context.Context, authorID, channelID uuid.UUID, content string, replyTo *uuid.UUID) (models.Message, error)
	GetMessages(ctx context.Context, channelID uuid.UUID, limit int, beforeID *uuid.UUID) ([]models.Message, error)
	DeleteMessage(ctx context.Context, channelID uuid.UUID, messageID uuid.UUID) error
}

type MessageBroker interface {
	PublishMessageCreated(ctx context.Context, msg models.Message) error
}

type MessageCache interface {
	GetChannelOwner(ctx context.Context, channelID uuid.UUID) (models.ChannelType, error)
}

type GuildProvider interface {
	CanWrite(ctx context.Context, userID, channelID uuid.UUID) (bool, error)
	CanRead(ctx context.Context, userID, channelID uuid.UUID) (bool, error)
}

type UserProvider interface {
	CanWrite(ctx context.Context, userID, channelID uuid.UUID) (bool, error)
	CanRead(ctx context.Context, userID, channelID uuid.UUID) (bool, error)
}

type MessageService struct {
	log    *zap.Logger
	repo   MessageRepo
	broker MessageBroker
	cache  MessageCache
	guilds GuildProvider
	users  UserProvider
}

func New(log *zap.Logger, repo MessageRepo, broker MessageBroker, cache MessageCache, guilds GuildProvider, users UserProvider) *MessageService {
	return &MessageService{
		log:    log,
		repo:   repo,
		broker: broker,
		cache:  cache,
		guilds: guilds,
		users:  users,
	}
}

func (ms *MessageService) SaveMessage(ctx context.Context, userID, channelID uuid.UUID, content string, replyTo *uuid.UUID) (*models.Message, error) {
	const op = "service.messages.SaveMessage"
	log := ms.log.With(zap.String("op", op))

	channelType, err := ms.cache.GetChannelOwner(ctx, channelID)
	if err != nil {
		log.Error("failed get channel service owner", zap.Error(err))
		return nil, err
	}

	var permission bool

	switch channelType {
	case models.ChannelTypeGuild:
		permission, err = ms.guilds.CanWrite(ctx, userID, channelID)
		if err != nil {
			log.Error("falied get permission form guilds", zap.Error(err))
			return nil, err
		}
	case models.ChannelTypeDM:
		permission, err = ms.users.CanWrite(ctx, userID, channelID)
		if err != nil {
			log.Error("failed get permission from users", zap.Error(err))
			return nil, err
		}
	default:
		log.Warn("unknown channel type", zap.String("type", string(channelType)))
		return nil, models.ErrInvalidChannel
	}

	if !permission {
		log.Debug("permission to write denied", zap.String("user_id", userID.String()), zap.String("channel_id", channelID.String()))
		return nil, models.ErrPermissionDenied
	}

	msg, err := ms.repo.SaveMessage(ctx, userID, channelID, content, replyTo)
	if err != nil {
		log.Error("failed save messsage", zap.Error(err))
		return nil, err
	}

	// По факту, если не получилось отправить в брокер это ошибка. Но из-за нее нельзя сообщать пользователю ошибку сохранния
	if err := ms.broker.PublishMessageCreated(ctx, msg); err != nil {
		log.Error("falied publich message to brocker", zap.Error(err))
	}

	return &msg, nil
}

func (ms *MessageService) GetMessages(ctx context.Context, userID, channelID uuid.UUID, limit int, beforeID *uuid.UUID) ([]models.Message, error) {
	const op = "service.messages.GetMessages"
	log := ms.log.With(zap.String("op", op))

	channelType, err := ms.cache.GetChannelOwner(ctx, channelID)
	if err != nil {
		if errors.Is(err, models.ErrChannelNotFound) {
			log.Debug("chanel not found in database", zap.String("channel_id", channelID.String()))
			return nil, err
		}
		log.Error("failed get channel service owner", zap.Error(err))
		return nil, err
	}

	var permission bool

	switch channelType {
	case models.ChannelTypeGuild:
		permission, err = ms.guilds.CanRead(ctx, userID, channelID)
		if err != nil {
			log.Error("failed get permission form guilds", zap.Error(err))
			return nil, err
		}
	case models.ChannelTypeDM:
		permission, err = ms.users.CanRead(ctx, userID, channelID)
		if err != nil {
			log.Error("failed get permission from users", zap.Error(err))
			return nil, err
		}
	default:
		log.Warn("unknown channel type", zap.String("type", string(channelType)))
		return nil, models.ErrInvalidChannel
	}

	if !permission {
		log.Debug("permission to read denied", zap.String("user_id", userID.String()), zap.String("channel_id", channelID.String()))
		return nil, models.ErrPermissionDenied
	}

	msg, err := ms.repo.GetMessages(ctx, channelID, limit, beforeID)
	if err != nil {
		if errors.Is(err, models.ErrChannelNotFound) {
			log.Debug("channel not found in cache", zap.String("channel_id", channelID.String()))
			return nil, err
		}
		log.Error("failed get messages from database", zap.Error(err))
		return nil, err
	}

	return msg, nil
}

// Best-политика удаления:
// 1. Получаем сообщеие с BD
// 2. (90%) Проверяем userID == authorID, удаляем сообщение
// 3. Для модерации в гильдии, если userID != authorID -> GuildProvider и проверяем права на удаление
// Я думаю что над реализацией удаления нужно еще подумать, к примеру было бы неплохо добавть time limit как в Telegram.
// Пока думаю для MVP Minor в 99% случаях этого будет достаточно
func (ms *MessageService) DeleteMessage(ctx context.Context, userID, channelID, messageID uuid.UUID) error {
}

// TODO: Implement later. Bulk deletion is heavy(in Cassandra), in Discord usess asynchronus soft-deletion
// func (ms *MessageService) DeleteAllMessage() {}
