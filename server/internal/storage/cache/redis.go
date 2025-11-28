package cache

import (
	"context"
	"time"

	"github.com/aerscs/theca-public/internal/config"
	"github.com/redis/go-redis/v9"
)

const (
	// PasswordResetTokenTTL время жизни токена сброса пароля (1 час)
	PasswordResetTokenTTL = time.Hour
)

// TokenCache интерфейс для работы с кэшем токенов
type TokenCache interface {
	// StoreResetToken сохраняет токен сброса пароля с TTL
	StoreResetToken(ctx context.Context, token string, userID uint) error
	// GetUserIDByResetToken получает ID пользователя по токену сброса пароля
	GetUserIDByResetToken(ctx context.Context, token string) (uint, error)
	// DeleteResetToken удаляет токен сброса пароля
	DeleteResetToken(ctx context.Context, token string) error
}

// redisTokenCache реализация интерфейса TokenCache
type redisTokenCache struct {
	client *redis.Client
}

// NewRedisTokenCache создаёт новый экземпляр редис-кэша
func NewRedisTokenCache(cfg *config.Config) TokenCache {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	return &redisTokenCache{
		client: client,
	}
}

// StoreResetToken сохраняет токен сброса пароля с TTL
func (r *redisTokenCache) StoreResetToken(ctx context.Context, token string, userID uint) error {
	return r.client.Set(ctx, getResetTokenKey(token), userID, PasswordResetTokenTTL).Err()
}

// GetUserIDByResetToken получает ID пользователя по токену сброса пароля
func (r *redisTokenCache) GetUserIDByResetToken(ctx context.Context, token string) (uint, error) {
	result, err := r.client.Get(ctx, getResetTokenKey(token)).Uint64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // токен не найден или истёк
		}
		return 0, err
	}
	return uint(result), nil
}

// DeleteResetToken удаляет токен сброса пароля
func (r *redisTokenCache) DeleteResetToken(ctx context.Context, token string) error {
	return r.client.Del(ctx, getResetTokenKey(token)).Err()
}

// getResetTokenKey возвращает ключ для токена сброса пароля
func getResetTokenKey(token string) string {
	return "password_reset:" + token
}
