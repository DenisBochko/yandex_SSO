package app

import (
	"context"
	grpcapp "yandex-sso/internal/app/grpc"
	"yandex-sso/internal/config"
	"yandex-sso/internal/services/auth"
	"yandex-sso/internal/services/users"
	postgresql "yandex-sso/internal/storage/postgres"
	miniostorage "yandex-sso/internal/storage/minio"
	minio "yandex-sso/pkg/minIO"
	"yandex-sso/pkg/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
	log        *zap.Logger
	dbConn     *pgxpool.Pool
}

func New(ctx context.Context, log *zap.Logger, cfg *config.Config) *App {
	// Создаём новый экземпляр подключения к бд
	conn, err := postgres.New(ctx, cfg.Postgres)

	if err != nil {
		log.Info("failed to connect to database", zap.Error(err))
		return nil
	}

	if conn.Ping(ctx) != nil {
		log.Info("failed to ping database", zap.Error(err))
		return nil
	}

	// Создаём новый экземпляр minIO клиента
	minioClient, err := minio.New(ctx, log, cfg.Minio)

	// Созадаём новый экземпляр хранилища postgresql
	postgresStorage := postgresql.New(conn)

	// Создаём новый экземпляр хранилища minIO
	minIOStorage := miniostorage.New(minioClient, cfg.Minio.Bucket)


	// Создаём новый экземпляр сервиса аутентификации
	authService := auth.New(log, postgresStorage, &cfg.Jwt)

	// Создаём новый экземпляр сервиса пользователей
	userService := users.New(log, postgresStorage, minIOStorage, &cfg.Jwt)

	// Создаём новый gRPC сервер
	// и регистрируем в нём сервисы аутентификации и пользователей
	grpcApp := grpcapp.New(log, authService, userService, cfg.GRPC.Port)

	return &App{
		GRPCServer: grpcApp,
		log:        log,
		dbConn:     conn,
	}
}

func (a *App) Stop() {
	a.GRPCServer.Stop()

	a.log.Info("stopping database connection")
	a.dbConn.Close()
}
