package user

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"os"
	"time"
)

// Service define the interface for the business logic of the User entity
type Service interface {
	CreateUser(ctx context.Context, u *User) error
	UpdateUser(ctx context.Context, u *User) error
	DeleteUser(ctx context.Context, id string) error
	GetUser(ctx context.Context, id string) (*User, error)
	ListUsers(ctx context.Context, filter *UserFilter) ([]User, int64, error)
}

// userService is the concrete implementation of the Service interface
type userService struct {
	repo   Repository
	logger *slog.Logger
}

func NewUserService(repo Repository) Service {
	handler := slog.NewTextHandler(os.Stdout, nil)
	return &userService{repo: repo, logger: slog.New(handler)}
}

// CreateUser create a user using the repository
func (s *userService) CreateUser(ctx context.Context, u *User) error {
	if u.Email == "" {
		return errors.New("email is required")
	}
	u.ID = uuid.New().String()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
	if err != nil {
		return err
	}
	u.Password = string(password)
	return s.repo.Create(ctx, u)
}

// UpdateUser update the user data and updated at timestamp
func (s *userService) UpdateUser(ctx context.Context, u *User) error {
	u.UpdatedAt = time.Now()
	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
	if err != nil {
		return err
	}
	u.Password = string(password)

	return s.repo.Update(ctx, u)
}

// DeleteUser delete a user by its ID
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteByID(ctx, id)
}

// GetUser gets a user by its ID
func (s *userService) GetUser(ctx context.Context, id string) (*User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ListUsers return the list of user according to filter
func (s *userService) ListUsers(ctx context.Context, filter *UserFilter) ([]User, int64, error) {
	return s.repo.List(ctx, filter)
}
