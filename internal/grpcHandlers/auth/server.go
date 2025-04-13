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

type Auth interface {
	Login(ctx context.Context, email string, password string) (string, string, error)
	Register(ctx context.Context, name string, email string, password string) (string, error)
	RefreshToken(ctx context.Context, token string) (string, error)
	Verify(ctx context.Context, token string) (bool, error)
}

type AuthServerAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &AuthServerAPI{auth: auth})
}

func (s *AuthServerAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	userID, err := s.auth.Register(ctx, req.GetUsername(), req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *AuthServerAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	accessToken, refreshToken, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())

	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServerAPI) RefreshToken(ctx context.Context, req *ssov1.RefreshTokenRequest) (*ssov1.RefreshTokenResponse, error) {
	if req.GetAccessToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	accessToken, err := s.auth.RefreshToken(ctx, req.GetAccessToken())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: req.GetAccessToken(), // возвращаем такой же refresh token
	}, nil
}

func (s *AuthServerAPI) Verify(ctx context.Context, req *ssov1.VerifyRequest) (*ssov1.VerifyResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
