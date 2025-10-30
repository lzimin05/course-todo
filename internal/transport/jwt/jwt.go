package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lzimin05/course-todo/config"
	errs "github.com/lzimin05/course-todo/internal/models/errs"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type Tokenator struct {
	sign          string
	tokenLifeSpan time.Duration
}

func NewTokenator(conf *config.JWTConfig) *Tokenator {
	return &Tokenator{
		sign:          conf.Signature,
		tokenLifeSpan: conf.TokenLifeSpan,
	}
}

func (t *Tokenator) CreateJWT(userID string) (string, error) {
	now := time.Now()
	expiration := now.Add(t.tokenLifeSpan)

	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.sign))
}

func (t *Tokenator) ParseJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(t.sign), nil
	})

	if err != nil {
		return nil, errs.ErrInvalidToken
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errs.ErrInvalidToken
}
