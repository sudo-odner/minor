package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/models"
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

// SetStatus Сохранить статус пользователя
func (c *Cache) SetStatus(ctx context.Context, userID string, status models.UserStatus, customStatus string, lastActiveAt int64) error {
	const op = "cache.redis.SetStatus"
	key := fmt.Sprintf("presence:user:%s", userID)

	err := c.client.HSet(ctx, key, map[string]interface{}{
		"status":         int32(status),
		"custom_status":  customStatus,
		"last_active_at": lastActiveAt,
	}).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetUserStatuses Получить статусы списка пользователей
func (c *Cache) GetUserStatuses(ctx context.Context, userIDs []string) (map[string]*models.Presence, error) {
	const op = "cache.redis.GetUserStatuses"
	if len(userIDs) == 0 {
		return make(map[string]*models.Presence), nil
	}

	pipe := c.client.Pipeline()
	cmds := make(map[string]*redis.MapStringStringCmd)

	for _, id := range userIDs {
		key := fmt.Sprintf("presence:user:%s", id)
		cmds[id] = pipe.HGetAll(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("%s: pipeline exec failed: %w", op, err)
	}

	res := make(map[string]*models.Presence)
	for id, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			continue
		}
		if len(val) == 0 {
			res[id] = &models.Presence{
				UserID:       id,
				Status:       models.UserStatusOffline,
				CustomStatus: "",
				LastActiveAt: 0,
			}
			continue
		}

		var status int32
		var lastActiveAt int64
		_, _ = fmt.Sscanf(val["status"], "%d", &status)
		_, _ = fmt.Sscanf(val["last_active_at"], "%d", &lastActiveAt)

		res[id] = &models.Presence{
			UserID:       id,
			Status:       models.UserStatus(status),
			CustomStatus: val["custom_status"],
			LastActiveAt: lastActiveAt,
		}
	}

	return res, nil
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
