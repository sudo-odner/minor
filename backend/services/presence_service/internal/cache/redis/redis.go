package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/config"
)

type Cache struct {
	client *redis.Client
}

// New Создание нового подключения к Redis
func New(ctx context.Context, cfg config.Redis) (*Cache, error) {
	const op = "cache.redis.New"

	// Setting
	opts, err := redis.ParseURL(cfg.Url)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse redis url: %w", op, err)
	}
	opts.PoolSize = cfg.PoolSize         // Максимальное колличество соединений на сервис
	opts.MinIdleConns = cfg.MinIdleConns // Минимальное значения откртых соединений (горячий старт)
	opts.DialTimeout = cfg.DialTimeout
	opts.ReadTimeout = cfg.ReadTimeout
	opts.WriteTimeout = cfg.WriteTimeout

	client := redis.NewClient(opts)

	// Check connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("%s: failed to ping redis: %w", op, err)
	}

	return &Cache{
		client: client,
	}, nil
}

// Ping Проверка соединения с Redis
func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Stop Закрыть соединение с Redis и освободить память
func (c *Cache) Stop() error {
	const op = "cache.redis.Stop"
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}
