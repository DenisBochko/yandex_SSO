package auth

import (
	"context"
	"errors"
	"yandex-sso/internal/services/auth"
	"yandex-sso/internal/storage"

	ssov1 "github.com/DenisBochko/yandex_contracts/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
// Login(context.Context, *LoginRequest) (*LoginResponse, error)
// RefreshToken(context.Context, *RefreshTokenRequest) (*RefreshTokenResponse, error)
// GetUserById(context.Context, *GetUserByIdRequest) (*GetUserByIdResponse, error)
// GetUsers(context.Context, *GetUsersRequest) (*GetUsersResponse, error)


type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (string, error)
	Register(ctx context.Context, email string, password string) (int64, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	userID, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.RegisterResponse{
		UserId: int64(userID),
	}, nil
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(-1))

	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.LoginResponse{
		AccessToken: token,
		RefreshToken: token,
	}, nil
}

func (s *serverAPI) RefreshToken(ctx context.Context, req *ssov1.RefreshTokenRequest) (*ssov1.RefreshTokenResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *serverAPI) GetUserById(ctx context.Context, req *ssov1.GetUserByIdRequest) (*ssov1.GetUserByIdResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *serverAPI) GetUsers(ctx context.Context, req *ssov1.GetUsersRequest) (*ssov1.GetUsersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
