package auth

import (
	"context"
	"errors"
	"fmt"
	"yandex-sso/internal/config"
	"yandex-sso/internal/domain/models"
	"yandex-sso/internal/storage"
	"yandex-sso/lib/jwt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Сервисный слой

type Auth struct {
	log       *zap.Logger
	storage   Storage
	cfg *config.JwtConfig
}

type Storage interface {
	SaveUser(ctx context.Context, name string, email string, passHash []byte) (uid string, err error)
	User(ctx context.Context, email string) (models.User, error)
}

func New(
	log *zap.Logger,
	storage Storage,
	cfg *config.JwtConfig,
) *Auth {
	return &Auth{
		log:       log,
		storage:   storage,
		cfg: cfg,
	}
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

// RegisterNewUser registers new user in the system and returns user ID.
// If user with given username already exists, returns error.
func (a *Auth) Register(ctx context.Context, name string, email string, pass string) (string, error) {
	log := a.log.With(zap.String("email", email))
	log.Info("Registering new user")

	// Генерируем хэш и соль для пароля.
	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", zap.Error(err))
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Сохраняем пользователя в БД
	id, err := a.storage.SaveUser(ctx, name, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Info("user already exists", zap.Error(err))
			return "", storage.ErrUserExists
		}

		log.Info("failed to save user", zap.Error(err))
		return "", fmt.Errorf("failed to save user: %w", err)
	}

	return id, nil
}

// Login checks if user with given credentials exists in the system and returns access token.
//
// If user exists, but password is incorrect, returns error.
// If user doesn't exist, returns error.
func (a *Auth) Login(ctx context.Context, email string, password string) (string, string, error) {
	log := a.log.With(zap.String("email", email))
	log.Info("Attempting to login user")

	// Достаем пользователя из БД
	user, err := a.storage.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", zap.Error(err))
			return "", "", fmt.Errorf("user not found: %w", ErrInvalidCredentials)
		}
		log.Error("failed to get user", zap.Error(err))
		return "", "", fmt.Errorf("failed to get user: %w", err)
	}

	// Проверяем корректность полученного пароля
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Warn("invalid credentials", zap.Error(err))
		return "", "", fmt.Errorf("invalid credentials: %w", ErrInvalidCredentials)
	}

	log.Info("user logged in successfully", zap.String("userID", user.ID))

	// Создаем access и refresh токен авторизации
	accessToken, err := jwt.NewToken(user, a.cfg.AppSecretAccessToken, a.cfg.AccessTokenTTL)
	if err != nil {
		log.Error("failed to create access token", zap.Error(err))
		return "", "", fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken, err := jwt.NewToken(user, a.cfg.AppSecretRefreshToken, a.cfg.RefreshTokenTTL)
	if err != nil {
		log.Error("failed to create refresh token", zap.Error(err))
		return "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (a *Auth) RefreshToken(ctx context.Context, token string) (string, error) {
	// Проверяем валидность токена
	claims, err := jwt.ValidateToken(token, a.cfg.AppSecretRefreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	user := models.User{
		ID: claims["uid"].(string),
		Email: claims["email"].(string),
	}

	// Создаем новый access токен
	newAccessToken, err := jwt.NewToken(user, a.cfg.AppSecretAccessToken, a.cfg.AccessTokenTTL)
	if err != nil {
		return "", fmt.Errorf("failed to create new token: %w", err)
	}

	return newAccessToken, nil
}

func (a *Auth) Verify(ctx context.Context, token string) (bool, error) {
	return true, nil
}