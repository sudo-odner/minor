package redis

import (
	"context"
	"fmt"
	"time"
)

const (
	resetTokenPrefix = "auth:reset:%s"
)

func (r *RedisSessionRepo) SetResetCode(ctx context.Context, email string, token string, ttl time.Duration) error {
	key := fmt.Sprintf(resetTokenPrefix, token)

	err := r.client.Set(ctx, key, email, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set reset code in redis: %w", err)
	}

	return nil
}

func (r *RedisSessionRepo) GetEmailByResetToken(ctx context.Context, token string) (string, error) {
    key := fmt.Sprintf(resetTokenPrefix, token) 

    res, err := r.client.Get(ctx, key).Result()
    if err != nil {
		return "", fmt.Errorf("failed to set reset code in redis: %w", err)
	}
   
	return res, nil
}

func (r *RedisSessionRepo) DeleteResetCode(ctx context.Context, token string) error {
	key := fmt.Sprintf(resetTokenPrefix, token)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete reset code from redis: %w", err)
	}

	return nil
}
