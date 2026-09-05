package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/models"
)

// GetChannelOwner get channel owner service (community or dm service)
func (c *Cache) GetChannelOwner(ctx context.Context, channelID uuid.UUID) (models.ChannelOwner, error) {
	const op = "cache.redis.GetChannelOwner"

	// 55 symbol * 1 bite (for one symbol) = 55 bite per write
	// ~ for 1.000.000 channel = 6.875MB in memory
	key := fmt.Sprintf("channel:%s:owner", channelID.String())

	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", models.ErrChannelNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return models.ChannelOwner(val), nil
}

// WriteChannelOwner write cahnnel owner in cache (TTL with 1 minute)
func (c *Cache) WriteChannelOwner(ctx context.Context, channelID uuid.UUID, owner models.ChannelOwner) error {
	const op = "cache.redis.WriteChannelOwner"

	key := fmt.Sprintf("channel:%s:owner", channelID.String())

	if err := c.client.Set(ctx, key, string(owner), 1*time.Minute).Err(); err != nil {
		return fmt.Errorf("falied to set channelOwner: %w", err)
	}

	return nil
}
