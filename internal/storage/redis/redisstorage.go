package redisstorage

import (
	"context"
	"fmt"
	"time"

	"github.com/DenisBochko/yandex_SSO/internal/storage"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
	ttl    time.Duration
}

func New(client *redis.Client, ttl time.Duration) *RedisStorage {
	return &RedisStorage{
		client: client,
		ttl:    ttl,
	}
}

func (r *RedisStorage) Set(uuid string, userID string) error {
	err := r.client.Set(context.Background(), uuid, userID, r.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set value in redis: %w", err)
	}

	return nil
}

func (r *RedisStorage) Get(uuid string) (string, error) {
	userID, err := r.client.Get(context.Background(), uuid).Result()
	if err == redis.Nil {
		return "", storage.ErrKeyDoesNotExist
	} else if err != nil {
		return "", storage.ErrInternalStorage
	}

	return userID, nil
}

func (r *RedisStorage) Delete(uuid string) error {
	err := r.client.Del(context.Background(), uuid).Err()
	if err != nil {
		return fmt.Errorf("failed to delete value from redis: %w", err)
	}
	return nil
}
