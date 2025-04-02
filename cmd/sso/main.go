package main

import (
	"yandex-sso/internal/config"
	"yandex-sso/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	cfg := config.MustLoad()

	logger := logger.SetupLogger(cfg.Env)
	defer logger.Sync()

	logger.Info("Starting SSO service",
		zap.Any("config", cfg),
	)

	logger.Debug("Debug message")
	// logger.Error("Error message",)
	// logger.Warn("Warning message",)
}
