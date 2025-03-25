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
	Password  string
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
	Create(context.Context, *User) error
	Update(context.Context, *User) error
	DeleteByID(context.Context, string) error
	GetByID(context.Context, string) (User, error)
	List(context.Context, *UserFilter) ([]User, int64, error)
	ExistsByEmail(context.Context, string) (bool, error)
}
