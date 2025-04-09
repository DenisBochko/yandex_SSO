package auth

import (
	"context"
	"errors"
	"fmt"
	"time"
	"yandex-sso/internal/domain/models"
	"yandex-sso/internal/storage"
	"yandex-sso/lib/jwt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Сервисный слой

type Auth struct {
	log         *zap.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserStorage interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
	User(ctx context.Context, email string) (models.User, error)
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

func New(
	log *zap.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		usrSaver:    userSaver,
		usrProvider: userProvider,
		log:         log,
		appProvider: appProvider,
		tokenTTL:    tokenTTL, // Время жизни возвращаемых токенов
	}
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

// Login checks if user with given credentials exists in the system and returns access token.
//
// If user exists, but password is incorrect, returns error.
// If user doesn't exist, returns error.
func (a *Auth) Login(ctx context.Context, email string, password string, appID int) (string, error) {
	log := a.log.With(zap.String("email", email))
	log.Info("Attempting to login user")

	// Достаем пользователя из БД
	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", zap.Error(err))
			return "", fmt.Errorf("user not found: %w", ErrInvalidCredentials)
		}

		log.Error("failed to get user", zap.Error(err))
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Проверяем корректность полученного пароля
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Warn("invalid credentials", zap.Error(err))
		return "", fmt.Errorf("invalid credentials: %w", ErrInvalidCredentials)
	}

	// Получаем информацию о приложении
	// app, err := a.appProvider.App(ctx, appID)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to get app: %w", err)
	// }
	app := models.App{
		ID:     appID,
		Name:   "test",
		Secret: "test",
	}

	log.Info("user logged in successfully", zap.Int("userID", int(user.ID)))

	// Создаем токен авторизации
	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to create token", zap.Error(err))
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	return token, nil
}

// RegisterNewUser registers new user in the system and returns user ID.
// If user with given username already exists, returns error.
func (a *Auth) Register(ctx context.Context, email string, pass string) (int64, error) {
	log := a.log.With(zap.String("email", email))
	log.Info("Registering new user")

	// Генерируем хэш и соль для пароля.
	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", zap.Error(err))
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	// Сохраняем пользователя в БД
	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", zap.Error(err))
			return 0, fmt.Errorf("user already exists: %w", storage.ErrUserExists)
		}

		log.Error("failed to save user", zap.Error(err))
		return 0, fmt.Errorf("failed to save user: %w", err)
	}

	return id, nil
}

// TODO: реализовать проверку прав администратора
func (a *Auth) IsAdmin(ctx context.Context, userID int) (bool, error) {
	// log := a.log.With(zap.Int("userID", userID))
	// log.Info("Checking if user is admin")

	// isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	// if err != nil {
	// 	log.Error("failed to get user", zap.Error(err))
	// 	return false, fmt.Errorf("failed to get user: %w", err)
	// }

	return false, nil
}
