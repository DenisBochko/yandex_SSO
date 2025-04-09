package app

import (
	"time"
	grpcapp "yandex-sso/internal/app/grpc"
	"yandex-sso/internal/services/auth"
	"yandex-sso/internal/storage/sqlite"

	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *zap.Logger, grpcPort int, storagePath string, tokenTTl time.Duration) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		log.Fatal("failed to create storage", zap.Error(err))
	}

	authService := auth.New(log, storage, storage, storage, tokenTTl)
	if err != nil {
		log.Fatal("failed to create auth service", zap.Error(err))
	}

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
