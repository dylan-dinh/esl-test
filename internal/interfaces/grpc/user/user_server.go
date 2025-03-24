package user

import (
	"context"
	"time"

	domainUser "github.com/dylan-dinh/esl-test/internal/domain/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserServer implements UserServiceServer and we inject the user service
type UserServer struct {
	UnimplementedUserServiceServer
	service domainUser.Service
}

// NewUserServer creates a new UserServer with the given service.
func NewUserServer(svc domainUser.Service) *UserServer {
	return &UserServer{service: svc}
}

// CreateUser is the RPC method to create a user
func (s *UserServer) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	newUser := &domainUser.User{
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		Nickname:  req.GetNickname(),
		Email:     req.GetEmail(),
		Country:   req.GetCountry(),
		Password:  req.GetPassword(),
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
	updatedUser := &domainUser.User{
		ID:        req.GetId(),
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
		Id:        req.GetId(),
		UpdatedAt: timestamppb.New(updatedUser.UpdatedAt),
	}, nil
}

// DeleteUser is the RPC method to delete a user
func (s *UserServer) DeleteUser(ctx context.Context, req *DeleteUserRequest) (*DeleteUserResponse, error) {
	if err := s.service.DeleteUser(ctx, req.GetId()); err != nil {
		return &DeleteUserResponse{}, err
	}
	return &DeleteUserResponse{
		Id: req.GetId(),
	}, nil
}

// GetUserById is the RPC method to get a user information by ID
func (s *UserServer) GetUserById(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
	var user *domainUser.User
	var err error

	if user, err = s.service.GetUser(ctx, req.GetId()); err != nil {
		return &GetUserResponse{}, err
	}

	return &GetUserResponse{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Country:   user.Country,
	}, nil
}
