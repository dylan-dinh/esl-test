package user

import (
	"context"
	"time"

	domainUser "github.com/dylan-dinh/esl-test/internal/domain/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserServer implements UserServiceServer
type UserServer struct {
	UnimplementedUserServiceServer
	service domainUser.Service // This is the user service injected
}

// NewUserServer creates a new UserServer with the given service.
func NewUserServer(svc domainUser.Service) *UserServer {
	return &UserServer{service: svc}
}

// CreateUser is an example gRPC method implementation.
func (s *UserServer) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	newUser := &domainUser.User{
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
		Id:        newUser.ID.String(),
		CreatedAt: timestamppb.New(newUser.CreatedAt),
	}, nil
}
