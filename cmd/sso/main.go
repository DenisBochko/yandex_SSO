package main

import (
	"os"
	"os/signal"
	"syscall"
	"yandex-sso/internal/app"
	"yandex-sso/internal/config"
	"yandex-sso/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	// кофигурация
	cfg := config.MustLoad()

	// инициализация логгера
	logger := logger.SetupLogger(cfg.Env)
	defer logger.Sync()

	logger.Info("Starting SSO service",
		zap.String("Environment", cfg.Env),
	)

	// инициализация приложения и его запуск
	application := app.New(logger, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	go application.GRPCServer.Run()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	logger.Info("Stopping SSO service...")

	application.GRPCServer.Stop()

	logger.Info("SSO service stopped")
}
