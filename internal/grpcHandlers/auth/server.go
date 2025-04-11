package auth

import (
	"context"
	"errors"
	"fmt"
	"yandex-sso/internal/domain/models"
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
	Login(ctx context.Context, email string, password string) (string, string, error)
	Register(ctx context.Context, name string, email string, password string) (string, error)
	RefreshToken(ctx context.Context, token string) (string, error)
	// GetUserById(ctx context.Context, id string) (string, error)
	GetUsers(ctx context.Context, ids []string) ([]models.User, error)
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

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
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

func (s *serverAPI) RefreshToken(ctx context.Context, req *ssov1.RefreshTokenRequest) (*ssov1.RefreshTokenResponse, error) {
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

func (s *serverAPI) GetUserById(ctx context.Context, req *ssov1.GetUserByIdRequest) (*ssov1.GetUserByIdResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *serverAPI) GetUsers(ctx context.Context, req *ssov1.GetUsersRequest) (*ssov1.GetUsersResponse, error) {
	if len(req.GetUserIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "ids is required")
	}

	fmt.Println("GetUsers", req.GetUserIds())

	users, err := s.auth.GetUsers(ctx, req.GetUserIds())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	var userResponses []*ssov1.User
	for _, user := range users {
		userResponses = append(userResponses,
			&ssov1.User{
				UserId: user.ID,
				Name:   user.Name,
				Email:  user.Email,
			})
	}

	return &ssov1.GetUsersResponse{
		Users: userResponses,
	}, nil
}
