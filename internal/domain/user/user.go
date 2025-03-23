package user

import (
	"context"
	"time"
)

// User domain will verify business logic if needed
type User struct {
	ID        string
	FirstName string
	LastName  string
	Nickname  string
	Email     string
	Country   string
	Password  string // hash it
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Filter is the filter by country
type Filter struct {
	Country string
}

// Repository define the interface to interact with the entity User
type Repository interface {
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	DeleteByID(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (User, error)
	List(ctx context.Context, filter *Filter) ([]User, error)
}
