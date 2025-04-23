package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DenisBochko/yandex_SSO/internal/config"
	"github.com/DenisBochko/yandex_SSO/internal/domain/models"
	"github.com/DenisBochko/yandex_SSO/internal/storage"
	"github.com/DenisBochko/yandex_SSO/lib/jwt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Сервисный слой

type Auth struct {
	log            *zap.Logger
	storage        Storage
	kafkaTransport KafkaTransport
	redis          RedisStorage
	cfg            *config.JwtConfig
}

type KafkaTransport interface {
	SendVerificationUserMessage(ctx context.Context, message models.VerificationUserMessage) error
}

type RedisStorage interface {
	Set(uuid string, userID string) error
	Get(uuid string) (string, error)
	Delete(uuid string) error
}

type Storage interface {
	SaveUser(ctx context.Context, name string, email string, passHash []byte) (uid string, err error)
	User(ctx context.Context, email string) (models.User, error)
	CreateVerificationToken(ctx context.Context, userID string, token string, expiresAt time.Time) (bool, error)
	VerifyToken(ctx context.Context, token string) (bool, error)
	UserById(ctx context.Context, id string) (models.User, error)
}

func New(
	log *zap.Logger,
	storage Storage,
	transport KafkaTransport,
	redis RedisStorage,
	cfg *config.JwtConfig,
) *Auth {
	return &Auth{
		log:            log,
		storage:        storage,
		kafkaTransport: transport,
		redis:          redis,
		cfg:            cfg,
	}
}

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserExists          = errors.New("user already exists")
	ErrRegistrationFailed  = errors.New("registration failed")
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
)

// apiGateway.com/api/sso/verify?token=edea549f-8843-492e-ad8e-c11a62e3bdc5

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

	// Создаем верификационный токен для пользователя
	verificationToken := uuid.NewString()
	expiresAt := time.Now().Add(10 * time.Minute).UTC()

	ok, err := a.storage.CreateVerificationToken(ctx, id, verificationToken, expiresAt)
	if err != nil {
		log.Error("failed to create verification token", zap.Error(err))
		return "", fmt.Errorf("failed to create verification token: %w", err)
	}

	if !ok {
		log.Error("failed to create verification token", zap.String("userID", id))
		return "", fmt.Errorf("failed to create verification token: %w", err)
	}

	// Отправляем сообщение в Kafka
	message := models.VerificationUserMessage{
		UserID: id,
		Name:   name,
		Email:  email,
		Token:  verificationToken,
	}

	if err := a.kafkaTransport.SendVerificationUserMessage(ctx, message); err != nil {
		log.Error("failed to send verification message", zap.Error(err))
		return "", fmt.Errorf("failed to send verification message: %w", err)
	}

	return id, nil
}

func (a *Auth) ResendVerificationToken(ctx context.Context, user_id string) (string, error) {
	user, err := a.storage.UserById(ctx, user_id)
	if err != nil {
		return "failed", err
	}

	// Создаем верификационный токен для пользователя
	verificationToken := uuid.NewString()
	expiresAt := time.Now().Add(10 * time.Minute).UTC()

	ok, err := a.storage.CreateVerificationToken(ctx, user.ID, verificationToken, expiresAt)
	if err != nil {
		a.log.Error("failed to create verification token", zap.Error(err))
		return "failed", fmt.Errorf("failed to create verification token: %w", err)
	}

	if !ok {
		a.log.Error("failed to create verification token", zap.String("userID", user.ID))
		return "failed", fmt.Errorf("failed to create verification token: %w", err)
	}

	// Отправляем сообщение в Kafka
	message := models.VerificationUserMessage{
		UserID: user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Token:  verificationToken,
	}

	if err := a.kafkaTransport.SendVerificationUserMessage(ctx, message); err != nil {
		a.log.Error("failed to send verification message", zap.Error(err))
		return "failed", fmt.Errorf("failed to send verification message: %w", err)
	}

	return "OK", nil
}

// Login checks if user with given credentials exists in the system and returns access token.
//
// If user exists, but password is incorrect, returns error.
// If user doesn't exist, returns error.
func (a *Auth) Login(ctx context.Context, email string, password string) (string, *timestamppb.Timestamp, string, *timestamppb.Timestamp, error) {
	log := a.log.With(zap.String("email", email))
	log.Info("Attempting to login user")

	// Достаем пользователя из БД
	user, err := a.storage.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", zap.Error(err))
			return "", nil, "", nil, fmt.Errorf("user not found: %w", ErrInvalidCredentials)
		}
		log.Error("failed to get user", zap.Error(err))
		return "", nil, "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Проверяем корректность полученного пароля
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Warn("invalid credentials", zap.Error(err))
		return "", nil, "", nil, fmt.Errorf("invalid credentials: %w", ErrInvalidCredentials)
	}

	log.Info("user logged in successfully", zap.String("userID", user.ID))

	// Создаем access токен авторизации
	accessToken, err := jwt.NewToken(user, a.cfg.AppSecretAccessToken, a.cfg.AccessTokenTTL)
	if err != nil {
		log.Error("failed to create access token", zap.Error(err))
		return "", nil, "", nil, fmt.Errorf("failed to create access token: %w", err)
	}
	access_token_expires_at := durationToTimestamp(time.Now(), a.cfg.AccessTokenTTL)

	// Создаем refresh токен авторизации и сохраняем в Redis
	refreshToken := uuid.New().String()

	if err := a.redis.Set(refreshToken, user.ID); err != nil {
		log.Error("failed to save user in redis", zap.Error(err))
		return "", nil, "", nil, fmt.Errorf("failed to save user in redis: %w", err)
	}
	refresh_token_expires_at := durationToTimestamp(time.Now(), a.cfg.RefreshTokenTTL)

	return accessToken, access_token_expires_at, refreshToken, refresh_token_expires_at, nil
}

func (a *Auth) RefreshToken(ctx context.Context, token string) (string, *timestamppb.Timestamp, string, *timestamppb.Timestamp, error) {
	userID, err := a.redis.Get(token)
	if err != nil {
		if errors.Is(err, storage.ErrKeyDoesNotExist) {
			return "", nil, "", nil, ErrRefreshTokenExpired
		}
		return "", nil, "", nil, fmt.Errorf("failed to get user from redis: %w", err)
	}

	user, err := a.storage.UserById(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Info("User not found")
			return "", nil, "", nil, storage.ErrUserNotFound
		}
		a.log.Info("failed to get user", zap.Error(err))
		return "", nil, "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	// удаляем старый токен из редис
	if err := a.redis.Delete(token); err != nil {
		return "", nil, "", nil, fmt.Errorf("failed to delete user from redis: %w", err)
	}

	// Создаем новый access токен
	newAccessToken, err := jwt.NewToken(user, a.cfg.AppSecretAccessToken, a.cfg.AccessTokenTTL)
	if err != nil {
		return "", nil, "", nil, fmt.Errorf("failed to create new token: %w", err)
	}
	access_token_expires_at := durationToTimestamp(time.Now(), a.cfg.AccessTokenTTL)

	// Создаем refresh токен авторизации и сохраняем в Redis
	refreshToken := uuid.New().String()

	if err := a.redis.Set(refreshToken, user.ID); err != nil {
		a.log.Error("failed to save user in redis", zap.Error(err))
		return "", nil, "", nil, fmt.Errorf("failed to save user in redis: %w", err)
	}
	refresh_token_expires_at := durationToTimestamp(time.Now(), a.cfg.RefreshTokenTTL)

	return newAccessToken, access_token_expires_at, refreshToken, refresh_token_expires_at, nil
}

func (a *Auth) Verify(ctx context.Context, token string) (bool, error) {
	log := a.log.With(zap.String("token", token))
	log.Info("Verifying token")

	ok, err := a.storage.VerifyToken(ctx, token)
	if err != nil {
		if errors.Is(err, storage.ErrTokenNotFound) {
			return false, fmt.Errorf("token not found: %w", err)
		}

		if errors.Is(err, storage.ErrTokenExpired) {
			return false, fmt.Errorf("token expired: %w", err)
		}
	}

	if !ok {
		return false, fmt.Errorf("token not found: %w", err)
	}

	return true, nil
}

func (a *Auth) Logut(ctx context.Context, refreshToken string) (bool, error) {
	_, err := a.redis.Get(refreshToken)
	if err != nil {
		if errors.Is(err, storage.ErrKeyDoesNotExist) {
			return false, ErrRefreshTokenExpired
		}
		return false, fmt.Errorf("failed to get user from redis: %w", err)
	}

	if err := a.redis.Delete(refreshToken); err != nil {
		a.log.Info("failed to delete user from redis")
		return false, fmt.Errorf("failed to delete user from redis: %w", err)
	}

	return true, nil
}

func durationToTimestamp(startTime time.Time, duration time.Duration) *timestamppb.Timestamp {
	t := startTime.Add(duration)
	return timestamppb.New(t)
}
