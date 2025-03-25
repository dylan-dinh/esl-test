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

var (
	ErrMissingEmailPassword = errors.New("email and password are required")
	ErrMissingName          = errors.New("first_name and last_name are required")
	ErrEmailExists          = errors.New("email already exists")
)

type Notifier interface {
	UserCreatedEvent(ctx context.Context, user *User) error
	UserUpdatedEvent(ctx context.Context, user *User) error
	UserDeletedEvent(ctx context.Context, id string) error
}

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
	mq     Notifier
	logger *slog.Logger
}

func NewUserService(repo Repository, mq Notifier) Service {
	handler := slog.NewTextHandler(os.Stdout, nil)
	return &userService{repo: repo, logger: slog.New(handler), mq: mq}
}

// CreateUser create a user using the repository
func (s *userService) CreateUser(ctx context.Context, u *User) error {
	if u.Email == "" || u.Password == "" {
		return ErrMissingEmailPassword
	}
	if u.FirstName == "" || u.LastName == "" {
		return ErrMissingName
	}
	exists, err := s.repo.ExistsByEmail(ctx, u.Email)
	if err != nil {
		return err
	}
	if exists {
		return ErrEmailExists
	}
	u.ID = uuid.New().String()
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
	if err != nil {
		return err
	}
	u.Password = string(password)

	if err = s.repo.Create(ctx, u); err != nil {
		return err
	}

	// fire and forget
	go func(u *User) {
		rabbitCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.mq.UserCreatedEvent(rabbitCtx, u); err != nil {
			slog.Error("failed to publish UserCreated event", "user_id", u.ID, "error", err)
		}
	}(u)

	return nil
}

// UpdateUser update the user data and updated at timestamp
func (s *userService) UpdateUser(ctx context.Context, u *User) error {
	u.UpdatedAt = time.Now()
	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
	if err != nil {
		return err
	}
	u.Password = string(password)

	go func(u *User) {
		rabbitCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.mq.UserUpdatedEvent(rabbitCtx, u); err != nil {
			slog.Error("failed to publish UserUpdated event", "user_id", u.ID, "error", err)
		}
	}(u)

	return s.repo.Update(ctx, u)
}

// DeleteUser delete a user by its ID
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	go func(id string) {
		rabbitCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.mq.UserDeletedEvent(rabbitCtx, id); err != nil {
			slog.Error("failed to publish UserUpdated event", "user_id", id, "error", err)
		}
	}(id)
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
