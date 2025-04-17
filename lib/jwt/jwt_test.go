package jwt

import (
	"testing"
	"time"

	jwtorigin "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/DenisBochko/yandex_SSO/internal/domain/models"
)

func TestNewToken(t *testing.T) {
	secret := "testsecret"
	duration := time.Hour

	user := models.User{
		ID:       "user-123",
		Email:    "test@example.com",
		Name:     "Test User",
		Verified: true,
		Avatar:   "avatar.png",
	}

	tokenStr, err := NewToken(user, secret, duration)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	// Парсим токен
	parsedToken, err := jwtorigin.Parse(tokenStr, func(token *jwtorigin.Token) (interface{}, error) {
		// Проверим, что используется ожидаемый метод подписи
		require.Equal(t, jwtorigin.SigningMethodHS256, token.Method)
		return []byte(secret), nil
	})
	require.NoError(t, err)
	require.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwtorigin.MapClaims)
	require.True(t, ok)

	require.Equal(t, user.ID, claims["uid"])
	require.Equal(t, user.Email, claims["email"])
	require.Equal(t, user.Name, claims["name"])
	require.Equal(t, user.Verified, claims["verified"])
	require.Equal(t, user.Avatar, claims["avatar"])

	expFloat, ok := claims["exp"].(float64)
	require.True(t, ok)
	expTime := time.Unix(int64(expFloat), 0)
	require.WithinDuration(t, time.Now().Add(duration), expTime, 2*time.Second)
}