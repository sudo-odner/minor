package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	resetCodePrefix    = "auth:reset:%s" 
)

func (r *RedisSessionRepo) SetResetCode(ctx context.Context, email string, code string, ttl time.Duration) error {
	key := fmt.Sprintf(resetCodePrefix, email)
	
	err := r.client.Set(ctx, key, code, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set reset code in redis: %w", err)
	}
	
	return nil
}

func (r *RedisSessionRepo) GetResetCode(ctx context.Context, email string) (string, error) {
	key := fmt.Sprintf(resetCodePrefix, email)
	
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("reset code not found or expired")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get reset code from redis: %w", err)
	}
	
	return val, nil
}

func (r *RedisSessionRepo) DeleteResetCode(ctx context.Context, email string) error {
	key := fmt.Sprintf(resetCodePrefix, email)
	
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete reset code from redis: %w", err)
	}
	
	return nil
}