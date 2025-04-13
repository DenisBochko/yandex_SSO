package users

import (
	"context"
	"errors"
	"yandex-sso/internal/domain/models"
	"yandex-sso/internal/storage"

	ssov1 "github.com/DenisBochko/yandex_contracts/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UsersService interface {
	GetUsers(ctx context.Context, ids []string) ([]models.User, error)
	GetUserById(ctx context.Context, id string) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) (bool, error)
	DeleteUser(ctx context.Context, id string) (bool, error)
}

type UsersServerAPI struct {
	ssov1.UnimplementedUsersServer
	userService UsersService
}

func Register(gRPC *grpc.Server, userService UsersService) {
	ssov1.RegisterUsersServer(gRPC, &UsersServerAPI{userService: userService})
}

func (u *UsersServerAPI) GetUserById(ctx context.Context, req *ssov1.GetUserByIdRequest) (*ssov1.GetUserByIdResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	user, err := u.userService.GetUserById(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.GetUserByIdResponse{
		User: &ssov1.User{
			UserId: user.ID,
			Name:   user.Name,
			Email:  user.Email,
		},
	}, nil
}

func (u *UsersServerAPI) GetUsers(ctx context.Context, req *ssov1.GetUsersRequest) (*ssov1.GetUsersResponse, error) {
	if len(req.GetUserIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "ids is required")
	}

	users, err := u.userService.GetUsers(ctx, req.GetUserIds())
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

func (s *UsersServerAPI) UpdateUser(ctx context.Context, req *ssov1.UpdateUserRequest) (*ssov1.UpdateUserResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}

	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "Name is required")
	}

	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "UserId is required")
	}

	user := models.User{
		ID:    req.GetUserId(),
		Name:  req.GetName(),
		Email: req.GetEmail(),
	}

	_, err := s.userService.UpdateUser(ctx, user)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.UpdateUserResponse{
		UserId: req.GetUserId(),
	}, nil
}

func (s *UsersServerAPI) DeleteUser(ctx context.Context, req *ssov1.DeleteUserRequest) (*ssov1.DeleteUserResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	ok, err := s.userService.DeleteUser(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !ok {
		return nil, status.Error(codes.Internal, "failed to delete user")
	}

	return &ssov1.DeleteUserResponse{
		UserId: req.GetUserId(),
	}, nil
}

func (s *UsersServerAPI) UploadPhoto(ctx context.Context, req *ssov1.UploadPhotoRequest) (*ssov1.UploadPhotoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
