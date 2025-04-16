package authinterceptor

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authInterceptor struct {
	appSecret string
}

type contextKey string

const (
	ContextUserIDKey   contextKey = "user_id"
	ContextEmailKey    contextKey = "email"
	ContextNameKey     contextKey = "name"
	ContextVerifiedKey contextKey = "verified"
	ContextAvatarKey   contextKey = "avatar"
)

func NewAuthInterceptor(appSecret string) (*authInterceptor, error) {
	if appSecret == "" {
		return nil, errors.New("appSecret cannot be empty")
	}
	return &authInterceptor{appSecret: appSecret}, nil
}

func (i *authInterceptor) UnaryAuthMiddleware(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// получение метаданных из контекста
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	// извлечение токена из метаданных
	token := md["authorization"]
	if len(token) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	// обрезаем преффикс Bearer, если он есть
	rawToken := token[0]
	if len(rawToken) > 7 && rawToken[:7] == "Bearer " {
		rawToken = rawToken[7:]
	}

	// проверка токена
	claims, err := i.validateToken(rawToken, i.appSecret)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("invalid token: %v", err))
	}

	// извлечение данных из токена
	userID, ok := claims["uid"].(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid uid in token")
	}
	email, ok := claims["email"].(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid email in token")
	}
	name, ok := claims["name"].(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid name in token")
	}
	verified, ok := claims["verified"].(bool)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid verified in token")
	}
	avatar, ok := claims["avatar"].(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid avatar in token")
	}
	// exp := claims["exp"].(float64) // JWT числовые значения возвращаются как float64

	// добавляем данные пользователя в контекст, так мы можем использовать их в последующих grpc методах
	ctx = context.WithValue(ctx, ContextUserIDKey, userID)
	ctx = context.WithValue(ctx, ContextEmailKey, email)
	ctx = context.WithValue(ctx, ContextNameKey, name)
	ctx = context.WithValue(ctx, ContextVerifiedKey, verified)
	ctx = context.WithValue(ctx, ContextAvatarKey, avatar)

	// вызываем следующий обработчик в цепочке
	return handler(ctx, req)
}

// ValidateToken проверяет валидность JWT токена и возвращает claims
func (i *authInterceptor) validateToken(tokenString string, appSecret string) (jwt.MapClaims, error) {
	// Парсим токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(appSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// Проверяем валидность токена
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
