package user

import (
	"context"
	"github.com/dylan-dinh/esl-test/internal/domain/user"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserServer implements UserServiceServer and we inject the user service
type UserServer struct {
	UnimplementedUserServiceServer
	service user.Service
}

// NewUserServer creates a new UserServer with the given service.
func NewUserServer(svc user.Service) *UserServer {
	return &UserServer{service: svc}
}

// CreateUser is the RPC method to create a user
func (s *UserServer) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	newUser := &user.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Nickname:  req.Nickname,
		Email:     req.Email,
		Country:   req.Country,
		Password:  req.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.service.CreateUser(ctx, newUser); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &CreateUserResponse{
		Id:        newUser.ID,
		CreatedAt: timestamppb.New(newUser.CreatedAt),
	}, nil
}

// UpdateUser is the RPC method to update a user information
func (s *UserServer) UpdateUser(ctx context.Context, req *UpdateUserRequest) (*UpdateUserResponse, error) {
	updatedUser := &user.User{
		ID:        req.Id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Nickname:  req.Nickname,
		Email:     req.Email,
		Country:   req.Country,
		Password:  req.Password,
	}

	if err := s.service.UpdateUser(ctx, updatedUser); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &UpdateUserResponse{
		Id:        req.Id,
		UpdatedAt: timestamppb.New(updatedUser.UpdatedAt),
	}, nil
}

// DeleteUser is the RPC method to delete a user
func (s *UserServer) DeleteUser(ctx context.Context, req *DeleteUserRequest) (*DeleteUserResponse, error) {
	if err := s.service.DeleteUser(ctx, req.Id); err != nil {
		return nil, err
	}
	return &DeleteUserResponse{
		Id: req.GetId(),
	}, nil
}

// GetUserById is the RPC method to get a user information by ID
func (s *UserServer) GetUserById(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
	var u *user.User
	var err error

	if u, err = s.service.GetUser(ctx, req.Id); err != nil {
		return nil, err
	}

	return &GetUserResponse{
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Nickname:  u.Nickname,
		Email:     u.Email,
		Country:   u.Country,
	}, nil
}

// ListUsers implements the ListUsers RPC.
func (s *UserServer) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	filter := &user.UserFilter{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Country:   req.Country,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}

	users, total, err := s.service.ListUsers(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	var pbUsers []*User
	for _, u := range users {
		pbUsers = append(pbUsers, &User{
			Id:        u.ID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Nickname:  u.Nickname,
			Email:     u.Email,
			Country:   u.Country,
			CreatedAt: timestamppb.New(u.CreatedAt),
			UpdatedAt: timestamppb.New(u.UpdatedAt),
		})
	}

	return &ListUsersResponse{
		Users:      pbUsers,
		TotalCount: total,
	}, nil
}
