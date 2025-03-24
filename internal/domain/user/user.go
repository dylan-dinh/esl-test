package user

import (
	"context"
	"time"
)

// User represent our entity user
type User struct {
	ID        string
	FirstName string `bson:"first_name"`
	LastName  string `bson:"last_name"`
	Nickname  string
	Email     string
	Country   string
	Password  string    // hash it
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// UserFilter holds criteria for filtering and paginating users
type UserFilter struct {
	FirstName string
	LastName  string
	Country   string
	Page      int32
	PageSize  int32
}

// Repository define the interface to interact with the entity User
type Repository interface {
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	DeleteByID(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (User, error)
	List(ctx context.Context, filter *UserFilter) ([]User, int64, error)
}
