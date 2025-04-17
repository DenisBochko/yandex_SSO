package redisClient

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisClientCfg struct {
	Host     string `yaml:"REDIS_HOST" env-required:"true"`
	Port     string `yaml:"REDIS_PORT" env-required:"true"`
	Password string `yaml:"REDIS_PASS" env-required:"true"`
	DB       int    `yaml:"REDIS_DB" env-required:"true"`
}

func New(ctx context.Context, log *zap.Logger, config RedisClientCfg) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	if response := rdb.Ping(ctx); response.Err() != nil {
		log.Error("failed to connect to redis", zap.Error(response.Err()))
		return nil
	}

	log.Info("connected to redis", zap.String("host", config.Host), zap.String("port", config.Port))
	return rdb
}
