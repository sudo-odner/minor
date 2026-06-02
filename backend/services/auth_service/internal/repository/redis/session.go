package redis

import (
	"context"
	"fmt"
	"time"
	"github.com/redis/go-redis/v9"
)

const (
	// Префикс для хранения: auth:rf:token_uuid -> user_id
	refreshPrefix = "auth:rf:%s"
	// Префикс для группировки сессий юзера (чтобы можно было удалить все сразу)
	userSessionsPrefix = "auth:usr:%s"
)

type RedisSessionRepo struct {
	client *redis.Client
}

func NewSessionRepo(client *redis.Client) *RedisSessionRepo {
	return &RedisSessionRepo{client: client}
}

// SetRefreshToken сохраняет UUID токена и связывает его с UserID
func (r *RedisSessionRepo) SetRefreshToken(ctx context.Context, userID string, tokenID string, ttl time.Duration) error {
	tokenKey := fmt.Sprintf(refreshPrefix, tokenID)
	userKey := fmt.Sprintf(userSessionsPrefix, userID)

	pipe := r.client.Pipeline()
	
	pipe.Set(ctx, tokenKey, userID, ttl)
	
	pipe.SAdd(ctx, userKey, tokenID)
	pipe.Expire(ctx, userKey, ttl)

	_, err := pipe.Exec(ctx)
	return err
}

// GetUserIDByRefreshToken возвращает userID, если токен существует
func (r *RedisSessionRepo) GetUserIDByRefreshToken(ctx context.Context, tokenID string) (string, error) {
	tokenKey := fmt.Sprintf(refreshPrefix, tokenID)
	userID, err := r.client.Get(ctx, tokenKey).Result()
	fmt.Println(tokenKey, userID)
	if err == redis.Nil {
		return "", fmt.Errorf("session not found")
	}
	return userID, err
}

// DeleteRefreshToken удаляет конкретную сессию (Logout)
func (r *RedisSessionRepo) DeleteRefreshToken(ctx context.Context, tokenID string) error {
	userID, err := r.GetUserIDByRefreshToken(ctx, tokenID)
	if err != nil {
		return nil
	}

	tokenKey := fmt.Sprintf(refreshPrefix, tokenID)
	userKey := fmt.Sprintf(userSessionsPrefix, userID)

	pipe := r.client.Pipeline()
	pipe.Del(ctx, tokenKey)
	pipe.SRem(ctx, userKey, tokenID)
	
	_, err = pipe.Exec(ctx)
	return err
}

// DeleteAllUserSessions удаляет все токены пользователя (при смене пароля)
func (r *RedisSessionRepo) DeleteAllUserSessions(ctx context.Context, userID string) error {
	userKey := fmt.Sprintf(userSessionsPrefix, userID)
	
	tokens, err := r.client.SMembers(ctx, userKey).Result()
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	for _, tID := range tokens {
		pipe.Del(ctx, fmt.Sprintf(refreshPrefix, tID))
	}
	pipe.Del(ctx, userKey)

	_, err = pipe.Exec(ctx)
	return err
}