package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type UserRepository interface {
	GetLastId(ctx context.Context) int
}

type AuthService struct {
	repo      UserRepository
	secretKey string
}

func NewAuthService(repo UserRepository, key string) *AuthService {
	return &AuthService{
		repo:      repo,
		secretKey: key,
	}
}

// Claims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const TOKEN_EXP = time.Hour * 24

// BuildJWTString создаёт токен и возвращает его в виде строки.
func (a *AuthService) BuildJWTString(ctx context.Context, id int) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		// собственное утверждение
		UserID: id,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(a.secretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func (a *AuthService) GetUserID(ctx context.Context, tokenString string) int {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(a.secretKey), nil
		})
	if err != nil {
		return -1
	}

	if !token.Valid {
		return -1
	}
	return claims.UserID
}

func (a *AuthService) GetLastId(ctx context.Context) int {
	return a.repo.GetLastId(ctx)
}
