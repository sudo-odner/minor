package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sudo-odner/minor/backend/services/chat_service/internal/config"
)

type Cache struct {
	client *redis.Client
}

func New(cfg config.Redis) (*Cache, error) {
	const op = "cache.redis.New"

	opts, err := redis.ParseURL(cfg.Url)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse redis url: %w", op, err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("%s: failed to ping redis: %w", op, err)
	}

	return &Cache{
		client: client,
	}, nil
}

func (c *Cache) Ping() error {
	return c.client.Ping(context.Background()).Err()
}

func (c *Cache) Stop() error {
	const op = "cache.redis.Stop"
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}
