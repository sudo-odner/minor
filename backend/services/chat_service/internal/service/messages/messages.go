package messages

import (
	"context"

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

func (ms *MessageService) SaveMessage(ctx context.Context, userID, channelID uuid.UUID, content string, replyTo *uuid.UUID) (models.Message, error) {
	// Get type channel
	// Get access write to channel

	// Save in Cassandra
	// Push in NATS
}

func (ms *MessageService) GetMessages(ctx context.Context, userID, channelID uuid.UUID, limit int, beforeID *uuid.UUID) ([]models.Message, error) {
	// Get type channel
	// Get access read to channel

	// Get Messages form Cassandra
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
