package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/aerscs/theca-public/internal/utils/errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	LoginRateLimit() gin.HandlerFunc
	PasswordResetRateLimit() gin.HandlerFunc
	EmailVerificationRateLimit() gin.HandlerFunc
}

type rateLimiter struct {
	redis *redis.Client
	log   *slog.Logger
}

func NewRateLimiter(redis *redis.Client, log *slog.Logger) RateLimiter {
	return &rateLimiter{
		redis: redis,
		log:   log,
	}
}

// LoginRateLimit ограничивает количество попыток входа с одного IP
func (rl *rateLimiter) LoginRateLimit() gin.HandlerFunc {
	return rl.createRateLimit("login", 5, time.Minute*15) // 5 попыток за 15 минут
}

// PasswordResetRateLimit ограничивает запросы на сброс пароля
func (rl *rateLimiter) PasswordResetRateLimit() gin.HandlerFunc {
	return rl.createRateLimit("password_reset", 3, time.Hour) // 3 попытки за час
}

// EmailVerificationRateLimit ограничивает отправку кодов верификации
func (rl *rateLimiter) EmailVerificationRateLimit() gin.HandlerFunc {
	return rl.createRateLimit("email_verification", 5, time.Minute*10) // 5 попыток за 10 минут
}

func (rl *rateLimiter) createRateLimit(keyPrefix string, maxAttempts int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "middleware.rateLimiter.createRateLimit"
		log := rl.log.With(slog.String("op", op))
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s:%s", keyPrefix, clientIP)

		ctx := context.Background()

		val, err := rl.redis.Get(ctx, key).Result()
		if err != nil && err != redis.Nil {
			log.Error("failed to get rate limit", slog.String("error", err.Error()))
			c.Next()
			return
		}

		attempts := 0
		if val != "" {
			attempts, _ = strconv.Atoi(val)
		}

		if attempts >= maxAttempts {
			log.Debug("too many requests", slog.String("client_ip", clientIP))
			errors.RespondWithError(c, errors.New(errors.CodeTooManyRequests, "Too many requests. Please try again later."))
			c.Abort()
			return
		}

		pipe := rl.redis.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		_, err = pipe.Exec(ctx)
		if err != nil {
			log.Error("failed to set rate limit", slog.String("error", err.Error()))
		}

		c.Next()
	}
}
