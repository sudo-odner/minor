package redis

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sudo-odner/minor/backend/services/chat_service/internal/models"
)

func (c *Cache) GetChannelOwner(ctx context.Context, channelID uuid.UUID) (models.ChannelType, error) {
	const op = "cache.redis.GetChannelOwner"

	// 55 symbol * 1 bite (for one symbol) = 55 bite per write
	// ~ for 1.000.000 channel = 6.875MB in memory
	key := fmt.Sprintf("channel:owner:%s", channelID.String())

	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", models.ErrChannelNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return models.ChannelType(val), nil
}
