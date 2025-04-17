package redisstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"yandex-sso/internal/domain/models"
	"yandex-sso/internal/storage"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
	ttl    time.Duration
}

func New(client *redis.Client, ttl time.Duration) *RedisStorage {
	return &RedisStorage{
		client: client,
		ttl: ttl,
	}
}

func (r *RedisStorage) Set(uuid string, user models.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal json")
	}

	err = r.client.Set(context.Background(), uuid, data, r.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set value in redis: %w", err)
	}

	return nil
}

func (r *RedisStorage) Get(uuid string) (*models.User, error) {
	data, err := r.client.Get(context.Background(), uuid).Bytes()
    if err == redis.Nil {
        return nil, storage.ErrKeyDoesNotExist
    } else if err != nil {
        return nil, storage.ErrInternalStorage
    }
    
	var user *models.User

	err = json.Unmarshal(data, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json")
	}

	return user, nil
}