package app

import (
	"context"
	"time"
	grpcapp "yandex-sso/internal/app/grpc"
	"yandex-sso/internal/services/auth"
	postgresql "yandex-sso/internal/storage/postgres"
	"yandex-sso/pkg/postgres"

	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(ctx context.Context, log *zap.Logger, grpcPort int, tokenTTl time.Duration, appSecret string, postgresCfg postgres.PostgresCfg) *App {
	// Подключение к БД
	conn, err := postgres.New(ctx, postgresCfg)

	if err != nil {
		log.Info("failed to connect to database", zap.Error(err))
		return nil
	}

	if conn.Ping(ctx) != nil {
		log.Info("failed to ping database", zap.Error(err))
		return nil
	}

	postgresStorage := postgresql.New(conn)
	// Создаём новый экземпляр сервиса аутентификации
	authService := auth.New(log, postgresStorage, tokenTTl, appSecret)

	// Создаём новый gRPC сервер
	// и регистрируем в нём сервис аутентификации
	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
