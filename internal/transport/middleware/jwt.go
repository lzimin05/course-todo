package middleware

import (
	"context"
	"net/http"

	"github.com/lzimin05/course-todo/internal/infrastructure/redis"
	"github.com/lzimin05/course-todo/internal/models/domains"
	"github.com/lzimin05/course-todo/internal/transport/jwt"
	response "github.com/lzimin05/course-todo/internal/transport/utils/response"
)

// AuthMiddleware создает middleware для проверки аутентификации и blacklist в Redis
func AuthMiddleware(tokenator *jwt.Tokenator, redisRepo *redis.AuthRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем токен из куки
			cookie, err := r.Cookie(domains.TokenCookieName)
			if err != nil {
				response.SendError(r.Context(), w, http.StatusUnauthorized, "Token cookie is required")
				return
			}

			// Парсим токен
			claims, err := tokenator.ParseJWT(cookie.Value)
			if err != nil {
				response.SendError(r.Context(), w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// Проверяем blacklist в Redis (если репозиторий передан)
			if redisRepo != nil {
				blacklisted, err := redisRepo.IsBlacklisted(r.Context(), claims.UserID, cookie.Value)
				if err != nil {
					response.SendError(r.Context(), w, http.StatusInternalServerError, "Internal server error")
					return
				}
				if blacklisted {
					response.SendError(r.Context(), w, http.StatusUnauthorized, "Token is blacklisted")
					return
				}
			}

			// Добавляем данные в контекст
			ctx := context.WithValue(r.Context(), domains.UserIDKey{}, claims.UserID) // Передаем запрос дальше
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
