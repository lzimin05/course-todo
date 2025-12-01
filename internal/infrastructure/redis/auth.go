package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/lzimin05/course-todo/config"
)

const (
	userTokensPrefix = "user_id:"
)

type AuthRepository struct {
	client *Client
	cfg    *config.JWTConfig
}

func NewAuthRepository(client *Client, jwtCfg *config.JWTConfig) *AuthRepository {
	return &AuthRepository{
		client: client,
		cfg:    jwtCfg,
	}
}

// AddToBlacklist добавляет токен в список недействительных токенов пользователя
func (r *AuthRepository) AddToBlacklist(ctx context.Context, userID, token string) error {
	expiration := time.Until(time.Now().Add(r.cfg.TokenLifeSpan))
	userKey := fmt.Sprintf("%s%s", userTokensPrefix, userID)

	// Добавляем токен в множество пользователя
	if err := r.client.SAdd(ctx, userKey, token).Err(); err != nil {
		return fmt.Errorf("failed to add token to user's blacklist: %w", err)
	}

	// Обновляем TTL для множества, чтобы после жизни самого старшего токена ключ очистился
	if err := r.client.Expire(ctx, userKey, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set expiration for user's blacklist: %w", err)
	}

	return nil
}

// IsBlacklisted проверяет, находится ли токен в черном списке пользователя
func (r *AuthRepository) IsBlacklisted(ctx context.Context, userID, token string) (bool, error) {
	userKey := fmt.Sprintf("%s%s", userTokensPrefix, userID)

	isMember, err := r.client.SIsMember(ctx, userKey, token).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token in user's blacklist: %w", err)
	}

	return isMember, nil
}
