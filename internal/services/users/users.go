package users

import (
	"context"
	"errors"
	"fmt"
	"yandex-sso/internal/config"
	"yandex-sso/internal/domain/models"
	"yandex-sso/internal/storage"

	"go.uber.org/zap"
)

type Storage interface {
	SaveUser(ctx context.Context, name string, email string, passHash []byte) (uid string, err error)
	User(ctx context.Context, email string) (models.User, error)
	Users(ctx context.Context, ids []string) ([]models.User, error)
	UserById(ctx context.Context, id string) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) (bool, error) // обновление по id
	DeleteUser(ctx context.Context, id string) (bool, error)
}

type UsersService struct {
	log     *zap.Logger
	storage Storage
	cfg     *config.JwtConfig
}

func New(
	log *zap.Logger,
	storage Storage,
	cfg *config.JwtConfig,
) *UsersService {
	return &UsersService{
		log:     log,
		storage: storage,
		cfg:     cfg,
	}
}

func (u *UsersService) GetUserById(ctx context.Context, id string) (models.User, error) {
	log := u.log.With(zap.String("id", id))
	log.Info("Getting user")

	user, err := u.storage.UserById(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Info("User not found")
			return models.User{}, storage.ErrUserNotFound
		}
		log.Info("failed to get user", zap.Error(err))
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (u *UsersService) GetUsers(ctx context.Context, ids []string) ([]models.User, error) {
	log := u.log.With(zap.String("ids", fmt.Sprintf("%v", ids)))
	log.Info("Getting users")

	users, err := u.storage.Users(ctx, ids)
	if err != nil {
		log.Info("failed to get users")

		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, storage.ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return users, nil
}

func (u *UsersService) UpdateUser(ctx context.Context, user models.User) (bool, error) {
	log := u.log.With(zap.String("id", user.ID))
	log.Info("Updating user")

	ok, err := u.storage.UpdateUser(ctx, user)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Info("User not found")
			return false, storage.ErrUserNotFound
		}
		log.Info("failed to update user", zap.Error(err))
		return false, fmt.Errorf("failed to update user: %w", err)
	}

	return ok, nil
}

func (u *UsersService) DeleteUser(ctx context.Context, id string) (bool, error) {
	log := u.log.With(zap.String("id", id))
	log.Info("Deleting user")

	ok, err := u.storage.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Info("User not found")
			return false, storage.ErrUserNotFound
		}
		log.Info("failed to delete user", zap.Error(err))
		return false, fmt.Errorf("failed to delete user: %w", err)
	}

	return ok, nil
}
