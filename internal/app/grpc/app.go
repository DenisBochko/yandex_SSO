package grpcapp

import (
	"fmt"
	"net"
	"github.com/DenisBochko/yandex_SSO/internal/config"
	grpcHandlersAuth "github.com/DenisBochko/yandex_SSO/internal/grpcHandlers/auth"
	grpcHandlersUsers "github.com/DenisBochko/yandex_SSO/internal/grpcHandlers/users"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	log        *zap.Logger
	gRPCServer *grpc.Server
	port       int
}

// Создаём новый gRPC сервер
func New(log *zap.Logger, cfg *config.Config, authService grpcHandlersAuth.Auth, userService grpcHandlersUsers.UsersService, port int) *App {
	// interceptor, err := authinterceptor.NewAuthInterceptor(cfg.Jwt.AppSecretAccessToken, []string{"/auth.Auth/Register", "/auth.Auth/Login"})
	// if err != nil {
	// 	log.Fatal("failed to create auth interceptor", zap.Error(err))
	// }

	// gRPCServer := grpc.NewServer(
	// 	grpc.UnaryInterceptor(interceptor.UnaryAuthMiddleware),
	// )

	gRPCServer := grpc.NewServer()

	grpcHandlersAuth.Register(gRPCServer, authService)
	grpcHandlersUsers.Register(gRPCServer, userService)

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
