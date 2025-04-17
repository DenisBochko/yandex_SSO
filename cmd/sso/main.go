package main

import (
	"context"
	"os/signal"
	"syscall"
	"github.com/DenisBochko/yandex_SSO/internal/app"
	"github.com/DenisBochko/yandex_SSO/internal/config"
	"github.com/DenisBochko/yandex_SSO/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	// Обработка сигналов завершения
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// кофигурация
	cfg := config.MustLoad()

	// инициализация логгера
	logger := logger.SetupLogger(cfg.Env)
	defer logger.Sync()

	logger.Info("Starting SSO service",
		zap.String("Environment", cfg.Env),
	)

	// инициализация приложения и его запуск
	application := app.New(ctx, logger, cfg)
	go application.GRPCServer.Run()

	// graceful shutdown
	// Ожидаем сигнал завершения
	<-ctx.Done()
	logger.Info("Stopping SSO service...")

	application.Stop()

	logger.Info("SSO service stopped")
}
