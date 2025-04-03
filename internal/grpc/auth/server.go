package auth

import (
	"context"

	ssov1 "github.com/DenisBochko/yandex_contracts/gen/go/sso"
	"google.golang.org/grpc"
)

// Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
// Login(context.Context, *LoginRequest) (*LoginResponse, error)
// IsAdmin(context.Context, *IsAdminRequest) (*IsAdminResponse, error)

type serverAPI struct {
	ssov1.UnimplementedAuthServer
}

func Register(gRPC *grpc.Server) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{})
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	return &ssov1.RegisterResponse{
		UserId: int64(123),
	}, nil
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	return &ssov1.LoginResponse{
		Token: "123",
	}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	return &ssov1.IsAdminResponse{
		IsAdmin: true,
	}, nil
}
