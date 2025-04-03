package app

import (
	"time"
	grpcapp "yandex-sso/internal/app/grpc"

	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *zap.Logger, grpcPort int, storagePath string, tokenTTl time.Duration) *App {
	grpcServer := grpcapp.New(log, grpcPort)

	return &App{
		GRPCServer: grpcServer,
	}
}
