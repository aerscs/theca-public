package repository

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	PasswordResetTokenTTL    = time.Hour
	EmailVerificationCodeTTL = time.Hour * 24
	FaviconCacheTTL          = time.Hour * 24 * 7
)

type CacheRepository interface {
	// ResetTokenCacheRepository defines interface for caching reset tokens
	ResetTokenCacheRepository
	// FaviconCacheRepository defines interface for caching favicon URLs
	FaviconCacheRepository
	// EmailVerificationCacheRepository defines interface for caching email verification code
	EmailVerificationCacheRepository
}

type ResetTokenCacheRepository interface {
	// StoreResetToken saves reset token with TTL
	StoreResetToken(ctx context.Context, token string, userID uint) error
	// GetUserIDByResetToken gets user ID by reset token
	GetUserIDByResetToken(ctx context.Context, token string) (uint, error)
	// DeleteResetToken deletes reset token
	DeleteResetToken(ctx context.Context, token string) error
}

type FaviconCacheRepository interface {
	// StoreFaviconURL saves favicon URL for the specified resource with TTL
	StoreFaviconURL(ctx context.Context, resourceURL, faviconURL string) error
	// GetFaviconURL returns favicon URL for the specified resource
	GetFaviconURL(ctx context.Context, resourceURL string) (string, error)
	// StoreFaviconBase64 saves favicon as base64 encoded string for the specified resource with TTL
	StoreFaviconBase64(ctx context.Context, resourceURL, faviconBase64 string) error
	// GetFaviconBase64 returns favicon as base64 encoded string for the specified resource
	GetFaviconBase64(ctx context.Context, resourceURL string) (string, error)
}

type EmailVerificationCacheRepository interface {
	// StoreEmailVerificationCode сохраняет код верификации email
	StoreEmailVerificationCode(ctx context.Context, userID uint, code string) error
	// GetEmailVerificationCode возвращает код верификации по ID пользователя
	GetEmailVerificationCode(ctx context.Context, userID uint) (string, error)
	// GetUserIDByVerificationCode возвращает ID пользователя по коду верификации
	GetUserIDByVerificationCode(ctx context.Context, code string) (uint, error)
	// DeleteEmailVerificationCode удаляет код верификации email
	DeleteEmailVerificationCode(ctx context.Context, userID uint) error
	// TrackVerificationAttempt отслеживает попытку ввода кода верификации
	TrackVerificationAttempt(ctx context.Context, userID uint) error
	// IsVerificationRateLimited проверяет превышение лимита попыток
	IsVerificationRateLimited(ctx context.Context, userID uint) (bool, error)
}

type redisRepository struct {
	client *redis.Client
	log    *slog.Logger
}

func NewRedisRepository(client *redis.Client, log *slog.Logger) CacheRepository {
	return &redisRepository{
		client: client,
		log:    log,
	}
}

// StoreResetToken saves reset token with TTL
func (r *redisRepository) StoreResetToken(ctx context.Context, token string, userID uint) error {
	const op = "redisRepository.StoreResetToken"
	log := r.log.With("op", op)

	err := r.client.Set(ctx, getResetTokenKey(token), userID, PasswordResetTokenTTL).Err()
	if err != nil {
		log.Error("failed to store reset token", "error", err, "token", token, "user_id", userID)
		return err
	}

	log.Debug("reset token stored", "token", token, "user_id", userID)
	return nil
}

// GetUserIDByResetToken gets user ID by reset token
func (r *redisRepository) GetUserIDByResetToken(ctx context.Context, token string) (uint, error) {
	const op = "redisRepository.GetUserIDByResetToken"
	log := r.log.With("op", op)

	result, err := r.client.Get(ctx, getResetTokenKey(token)).Uint64()
	if err != nil {
		if err == redis.Nil {
			log.Debug("token not found or expired", "token", token)
			return 0, nil // token not found or expired
		}
		log.Error("failed to get user ID by reset token", "error", err, "token", token)
		return 0, err
	}

	log.Debug("user ID retrieved by reset token", "token", token, "user_id", result)
	return uint(result), nil
}

// DeleteResetToken deletes reset token
func (r *redisRepository) DeleteResetToken(ctx context.Context, token string) error {
	const op = "redisRepository.DeleteResetToken"
	log := r.log.With("op", op)

	err := r.client.Del(ctx, getResetTokenKey(token)).Err()
	if err != nil {
		log.Error("failed to delete reset token", "error", err, "token", token)
		return err
	}

	log.Debug("reset token deleted", "token", token)
	return nil
}

// StoreEmailVerificationCode сохраняет код верификации email в Redis
func (r *redisRepository) StoreEmailVerificationCode(ctx context.Context, userID uint, code string) error {
	const op = "redisRepository.StoreEmailVerificationCode"
	log := r.log.With("op", op)

	// Сохраняем код по userID для получения кода пользователя
	err := r.client.Set(ctx, getEmailVerificationCodeKey(userID), code, EmailVerificationCodeTTL).Err()
	if err != nil {
		log.Error("failed to store email verification code by userID", "error", err, "user_id", userID)
		return err
	}

	// Сохраняем userID по коду для поиска пользователя по коду
	err = r.client.Set(ctx, getEmailVerificationUserKey(code), userID, EmailVerificationCodeTTL).Err()
	if err != nil {
		log.Error("failed to store userID by email verification code", "error", err, "code", code)
		return err
	}

	log.Debug("email verification code stored", "user_id", userID, "code", code)
	return nil
}

// GetEmailVerificationCode возвращает код верификации по ID пользователя
func (r *redisRepository) GetEmailVerificationCode(ctx context.Context, userID uint) (string, error) {
	const op = "redisRepository.GetEmailVerificationCode"
	log := r.log.With("op", op)

	code, err := r.client.Get(ctx, getEmailVerificationCodeKey(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			log.Debug("verification code not found", "user_id", userID)
			return "", nil
		}
		log.Error("failed to get verification code", "error", err, "user_id", userID)
		return "", err
	}

	log.Debug("verification code retrieved", "user_id", userID, "code", code)
	return code, nil
}

// GetUserIDByVerificationCode возвращает ID пользователя по коду верификации
func (r *redisRepository) GetUserIDByVerificationCode(ctx context.Context, code string) (uint, error) {
	const op = "redisRepository.GetUserIDByVerificationCode"
	log := r.log.With("op", op)

	result, err := r.client.Get(ctx, getEmailVerificationUserKey(code)).Uint64()
	if err != nil {
		if err == redis.Nil {
			log.Debug("verification code not found or expired", "code", code)
			return 0, nil
		}
		log.Error("failed to get user ID by verification code", "error", err, "code", code)
		return 0, err
	}

	log.Debug("user ID retrieved by verification code", "code", code, "user_id", result)
	return uint(result), nil
}

// DeleteEmailVerificationCode удаляет код верификации email
func (r *redisRepository) DeleteEmailVerificationCode(ctx context.Context, userID uint) error {
	const op = "redisRepository.DeleteEmailVerificationCode"
	log := r.log.With("op", op)

	// Сначала получаем код
	code, err := r.GetEmailVerificationCode(ctx, userID)
	if err != nil {
		return err
	}
	if code == "" {
		return nil // Код не найден, нечего удалять
	}

	// Удаляем код по userID
	err = r.client.Del(ctx, getEmailVerificationCodeKey(userID)).Err()
	if err != nil {
		log.Error("failed to delete email verification code", "error", err, "user_id", userID)
		return err
	}

	// Удаляем userID по коду
	err = r.client.Del(ctx, getEmailVerificationUserKey(code)).Err()
	if err != nil {
		log.Error("failed to delete user ID by verification code", "error", err, "code", code)
		return err
	}

	log.Debug("email verification code deleted", "user_id", userID, "code", code)
	return nil
}

// StoreFaviconURL saves favicon URL with TTL
func (r *redisRepository) StoreFaviconURL(ctx context.Context, resourceURL, faviconURL string) error {
	const op = "redisRepository.StoreFaviconURL"
	log := r.log.With("op", op)

	err := r.client.Set(ctx, getFaviconKey(resourceURL), faviconURL, FaviconCacheTTL).Err()
	if err != nil {
		log.Error("failed to store favicon URL", "error", err, "resource_url", resourceURL)
		return err
	}

	log.Debug("favicon URL stored", "resource_url", resourceURL, "favicon_url", faviconURL)
	return nil
}

// GetFaviconURL returns favicon URL
func (r *redisRepository) GetFaviconURL(ctx context.Context, resourceURL string) (string, error) {
	const op = "redisRepository.GetFaviconURL"
	log := r.log.With("op", op)

	faviconURL, err := r.client.Get(ctx, getFaviconKey(resourceURL)).Result()
	if err != nil {
		if err == redis.Nil {
			log.Debug("favicon URL not found in cache", "resource_url", resourceURL)
			return "", nil
		}
		log.Error("failed to get favicon URL", "error", err, "resource_url", resourceURL)
		return "", err
	}

	log.Debug("favicon URL retrieved from cache", "resource_url", resourceURL, "favicon_url", faviconURL)
	return faviconURL, nil
}

// StoreFaviconBase64 saves favicon base64 data with TTL
func (r *redisRepository) StoreFaviconBase64(ctx context.Context, resourceURL, faviconBase64 string) error {
	const op = "redisRepository.StoreFaviconBase64"
	log := r.log.With("op", op)

	err := r.client.Set(ctx, getFaviconBase64Key(resourceURL), faviconBase64, FaviconCacheTTL).Err()
	if err != nil {
		log.Error("failed to store favicon base64", "error", err, "resource_url", resourceURL)
		return err
	}

	log.Debug("favicon base64 stored", "resource_url", resourceURL)
	return nil
}

// GetFaviconBase64 returns favicon base64 data
func (r *redisRepository) GetFaviconBase64(ctx context.Context, resourceURL string) (string, error) {
	const op = "redisRepository.GetFaviconBase64"
	log := r.log.With("op", op)

	faviconBase64, err := r.client.Get(ctx, getFaviconBase64Key(resourceURL)).Result()
	if err != nil {
		if err == redis.Nil {
			log.Debug("favicon base64 not found in cache", "resource_url", resourceURL)
			return "", nil
		}
		log.Error("failed to get favicon base64", "error", err, "resource_url", resourceURL)
		return "", err
	}

	log.Debug("favicon base64 retrieved from cache", "resource_url", resourceURL)
	return faviconBase64, nil
}

func (r *redisRepository) TrackVerificationAttempt(ctx context.Context, userID uint) error {
	const op = "redisRepository.TrackVerificationAttempt"
	log := r.log.With("op", op)

	key := getVerificationAttemptsKey(userID)

	pipe := r.client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Hour)
	_, err := pipe.Exec(ctx)

	if err != nil {
		log.Error("failed to track verification attempt", "error", err, "user_id", userID)
		return err
	}

	log.Debug("verification attempt tracked", "user_id", userID)
	return nil
}

func (r *redisRepository) IsVerificationRateLimited(ctx context.Context, userID uint) (bool, error) {
	const op = "redisRepository.IsVerificationRateLimited"
	log := r.log.With("op", op)

	key := getVerificationAttemptsKey(userID)
	attempts, err := r.client.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		log.Error("failed to check verification rate limit", "error", err, "user_id", userID)
		return false, err
	}

	const maxAttempts = 5
	isLimited := attempts >= maxAttempts

	if isLimited {
		log.Debug("verification rate limited", "user_id", userID, "attempts", attempts)
	}

	return isLimited, nil
}

// getResetTokenKey returns key for reset token
func getResetTokenKey(token string) string {
	return "password_reset:" + token
}

// getEmailVerificationCodeKey returns key for storing email verification code by userID
func getEmailVerificationCodeKey(userID uint) string {
	return "email_verification:user:" + strconv.FormatUint(uint64(userID), 10)
}

// getEmailVerificationUserKey returns key for storing userID by email verification code
func getEmailVerificationUserKey(code string) string {
	return "email_verification:code:" + code
}

// getFaviconKey returns key for storing favicon URL
func getFaviconKey(resourceURL string) string {
	return "favicon:" + resourceURL
}

// getFaviconBase64Key returns key for storing favicon base64 data
func getFaviconBase64Key(resourceURL string) string {
	return "favicon_base64:" + resourceURL
}

// getVerificationAttemptsKey returns key for storing verification attempts
func getVerificationAttemptsKey(userID uint) string {
	return "verification_attempts:user:" + strconv.FormatUint(uint64(userID), 10)
}
