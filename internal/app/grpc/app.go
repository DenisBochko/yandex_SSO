package grpcapp

import (
	"fmt"
	"net"
	grpcHandlersAuth "yandex-sso/internal/grpcHandlers/auth"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	log        *zap.Logger
	gRPCServer *grpc.Server
	port       int
}

// Создаём новый gRPC сервер
func New(log *zap.Logger, authService grpcHandlersAuth.Auth, port int) *App {
	gRPCServer := grpc.NewServer()
	grpcHandlersAuth.Register(gRPCServer, authService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("failed to listen tcp: %v", err)
	}

	a.log.Info("gRPC server is running", zap.Int("port", a.port), zap.String("addres", lis.Addr().String()))

	// Запускаем сервер на порту
	if err := a.gRPCServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

func (a *App) Stop() {
	a.log.Info("stopping gRPC server", zap.Int("port", a.port))
	
	a.gRPCServer.GracefulStop()
}
