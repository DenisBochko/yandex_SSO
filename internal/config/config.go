package config

import (
	"flag"
	"os"
	"time"
	"github.com/DenisBochko/yandex_SSO/pkg/kafka"
	minio "github.com/DenisBochko/yandex_SSO/pkg/minIO"
	"github.com/DenisBochko/yandex_SSO/pkg/postgres"
	redisClient "github.com/DenisBochko/yandex_SSO/pkg/redis"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string                     `yaml:"env" env-default:"local"`
	Jwt      JwtConfig                  `yaml:"jwt"`
	GRPC     GRPCConfig                 `yaml:"grpc"`
	Postgres postgres.PostgresCfg       `yaml:"POSTGRES"`
	Minio    minio.MinioConfig          `yaml:"MINIO"`
	Kafka    kafka.KafkaConfig          `yaml:"KAFKA"`
	Redis    redisClient.RedisClientCfg `yaml:"REDIS"`
}

type JwtConfig struct {
	AppSecretAccessToken  string        `yaml:"app_secret_a" env-required:"true"`
	AppSecretRefreshToken string        `yaml:"app_secret_b" env-required:"true"`
	AccessTokenTTL        time.Duration `yaml:"access_token_ttl" env-required:"true"`
	RefreshTokenTTL       time.Duration `yaml:"refresh_token_ttl" env-required:"true"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

// Must - значит, что функция не возвращает ошибку, а паникует, если не удалось загрузить конфигурацию
func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	// Проверяем, что файл существует
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	var config Config

	if err := cleanenv.ReadConfig(path, &config); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &config
}

// fetchConfigPath - получает путь к конфигурации из переменной окружения или флага при запуске
// Приоритет: флаг > переменная окружения > значение по умолчанию
// Если не удалось получить путь, возвращает пустую строку
func fetchConfigPath() string {
	var result string

	flag.StringVar(&result, "config", "", "Path to config file")
	flag.Parse()

	if result == "" {
		result = os.Getenv("CONFIG_PATH")
	}

	return result
}
