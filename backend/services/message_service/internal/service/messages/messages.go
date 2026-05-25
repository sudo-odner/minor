package messages

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/models"
	"go.uber.org/zap"
)

type MessageRepo interface {
	SaveMessage(ctx context.Context, authorID, channelID uuid.UUID, content string, replyTo *uuid.UUID) (*models.Message, error)
	GetMessages(ctx context.Context, channelID uuid.UUID, limit int, beforeID *uuid.UUID) ([]models.Message, error)
	GetMessage(ctx context.Context, channelID, messageID uuid.UUID) (*models.Message, error)
	DeleteMessage(ctx context.Context, channelID uuid.UUID, messageID uuid.UUID) error
}

type MessageBroker interface {
	PublishMessageCreated(ctx context.Context, msg models.Message) error
	PublishMessageDeleted(ctx context.Context, channelID, messageID uuid.UUID) error
}

type MessageCache interface {
	GetChannelOwner(ctx context.Context, channelID uuid.UUID) (models.ChannelOwner, error)
	WriteChannelOwner(ctx context.Context, channelID uuid.UUID, owner models.ChannelOwner) error
}

type CommunityClient interface {
	FetchPermission(ctx context.Context, userID, channelID uuid.UUID) (authz.Permission, error)
	CheckChannelExists(ctx context.Context, channelID uuid.UUID) (bool, error)
}

type UserClient interface {
	FetchPermission(ctx context.Context, userID, channelID uuid.UUID) (authz.Permission, error)
	CheckChannelExists(ctx context.Context, channelID uuid.UUID) (bool, error)
}

type MessageService struct {
	log             *zap.Logger
	repo            MessageRepo
	broker          MessageBroker
	cache           MessageCache
	communityClient CommunityClient
	userClient      UserClient
}

func New(log *zap.Logger, repo MessageRepo, broker MessageBroker, cache MessageCache, communityClient CommunityClient, userClient UserClient) *MessageService {
	return &MessageService{
		log:             log,
		repo:            repo,
		broker:          broker,
		cache:           cache,
		communityClient: communityClient,
		userClient:      userClient,
	}
}

// TODO: Сделать нормальное логирование

// loadChennelOwner Получить кто владеет channelID (TODO: перелать метод, messages_service не должен этим заниматься)
func (ms *MessageService) loadChannelOwner(
	ctx context.Context,
	channelID uuid.UUID,
) (models.ChannelOwner, error) {
	const op = "service.messages.loadChannelOwner"

	// 1. Пытаемся получить данные из кеша
	channelOwner, err := ms.cache.GetChannelOwner(ctx, channelID)
	if err == nil {
		return channelOwner, nil
	}

	// 2. Кеш промах (Cache miss). Вручную опрашивем сервисы.
	// 2.1 Опрашиваем Community service
	ok, err := ms.communityClient.CheckChannelExists(ctx, channelID)
	if err != nil {
		// TODO: проверка на то что сеть упала/timeout
		return "", fmt.Errorf("%s (community grpc): %w", op, err)
	}
	if ok {
		_ = ms.cache.WriteChannelOwner(ctx, channelID, models.ChannelOwnerCommunity)
		return models.ChannelOwnerCommunity, nil
	}

	// 2.2 Опрашиваем User service
	ok, err = ms.userClient.CheckChannelExists(ctx, channelID)
	if err != nil {
		// TODO: проверка на то что сеть упала/timeout
		return "", fmt.Errorf("%s (user grpc): %w", op, err)
	}
	if ok {
		_ = ms.cache.WriteChannelOwner(ctx, channelID, models.ChannelOwnerUser)
		return models.ChannelOwnerUser, nil
	}

	return "", fmt.Errorf("%s: not found owner channelID: %w", op, models.ErrChannelNotFound)
}

// loadPermissionMask Получить маску доступов
func (ms *MessageService) loadPermissionMask(ctx context.Context, userID, channelID uuid.UUID) (authz.Permission, error) {
	const op = "service.messages.loadPermission"

	// 1. Получаем владельца channelID
	channelOwner, err := ms.loadChannelOwner(ctx, channelID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// 2. Идем в нуный сервис, чтобы получить макску прав
	var maskPermission authz.Permission
	switch channelOwner {
	case models.ChannelOwnerCommunity:
		mask, err := ms.communityClient.FetchPermission(ctx, userID, channelID)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
		maskPermission = mask
	case models.ChannelOwnerUser:
		mask, err := ms.communityClient.FetchPermission(ctx, userID, channelID)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
		maskPermission = mask
	}
	return maskPermission, nil
}

// Сохранение сообщения
func (ms *MessageService) SaveMessage(
	ctx context.Context,
	userID, channelID uuid.UUID,
	content string,
	replyTo *uuid.UUID,
) (*models.Message, error) {
	const op = "service.messages.SaveMessage"
	log := ms.log.With(zap.String("op", op))

	// Получаем максу доступов
	maskPermission, err := ms.loadPermissionMask(ctx, userID, channelID)
	if err != nil {
		if errors.Is(err, models.ErrChannelNotFound) {
			log.Debug("owner channelID not found", zap.String("channelID", channelID.String()))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		log.Error("falied to resolve channel owner", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Проверка прав доступа на запись
	if !authz.Has(maskPermission, authz.PermSendMessages) {
		log.Debug(
			"permission denied to save message in channel",
			zap.String("userID", userID.String()),
			zap.String("channelID", channelID.String()),
		)
		return nil, models.ErrPermissionDenied
	}

	// Сохранение сообщения
	msg, err := ms.repo.SaveMessage(ctx, userID, channelID, content, replyTo)
	if err != nil {
		log.Error("failed save messsage", zap.Error(err))
		return nil, err
	}

	if err := ms.broker.PublishMessageCreated(ctx, *msg); err != nil {
		log.Error("failed publish message to broker", zap.Error(err))
	}

	return msg, nil
}

// GetMessages Получить сообщение
func (ms *MessageService) GetMessages(
	ctx context.Context,
	userID, channelID uuid.UUID,
	limit int,
	beforeID *uuid.UUID,
) ([]models.Message, error) {
	const op = "service.messages.GetMessages"
	log := ms.log.With(zap.String("op", op))

	// Получаем максу доступов
	maskPermission, err := ms.loadPermissionMask(ctx, userID, channelID)
	if err != nil {
		if errors.Is(err, models.ErrChannelNotFound) {
			log.Debug("owner channelID not found", zap.String("channelID", channelID.String()))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		log.Error("falied to resolve channel owner", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Проверка прав доступа на запись
	if !authz.Has(maskPermission, authz.PermReadChat) {
		log.Debug(
			"permission denied to read message in channel",
			zap.String("userID", userID.String()),
			zap.String("channelID", channelID.String()),
		)
		return nil, models.ErrPermissionDenied
	}

	// Получение сообщения
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
// 1. [+] Получаем сообщеие с BD
// 2. [+] (90%) Проверяем userID == authorID, удаляем сообщение
// 3. [ ] Для модерации в гильдии, если userID != authorID -> GuildProvider и проверяем права на удаление
// 4. [+] Посылаем в шину для удаления соощения у остальных
//
// Я думаю что над реализацией удаления нужно еще подумать, к примеру было бы неплохо добавть time limit как в Telegram.
// Пока думаю для MVP Minor в 99% случаях этого будет достаточно
func (ms *MessageService) DeleteMessage(ctx context.Context, userID, channelID, messageID uuid.UUID) error {
	const op = "service.messages.DeleteMessage"
	log := ms.log.With(zap.String("op", op))

	msg, err := ms.repo.GetMessage(ctx, channelID, messageID)
	if err != nil {
		if errors.Is(err, models.ErrMessageNotFound) {
			log.Debug("not found message", zap.String("message_id", messageID.String()))
			return err
		}
		log.Error("failed get message from database", zap.Error(err))
		return err
	}

	if msg.UserID != userID {
		log.Debug("permission to delete denied", zap.String("user_id", userID.String()), zap.String("message_id", messageID.String()))
		return models.ErrPermissionDenied
	}

	if err = ms.repo.DeleteMessage(ctx, channelID, messageID); err != nil {
		if errors.Is(err, models.ErrMessageNotFound) {
			log.Debug("not found message", zap.String("message_id", messageID.String()))
			return err
		}
		log.Error("failed delete message", zap.Error(err))
		return err
	}

	if err = ms.broker.PublishMessageDeleted(ctx, channelID, messageID); err != nil {
		log.Error("failed to publish delete to broker", zap.Error(err))
	}

	return nil
}

// TODO: Implement later. Bulk deletion is heavy(in Cassandra), in Discord usess asynchronus soft-deletion
// func (ms *MessageService) DeleteAllMessage() {}
