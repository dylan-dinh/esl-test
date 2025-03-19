package repository

import (
	"context"
	"errors"
	"github.com/dylan-dinh/esl-test/internal/domain/user"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log/slog"
	"os"
)

// UserRepository concrete implementation of user.Repository
type UserRepository struct {
	coll   *mongo.Collection
	logger *slog.Logger
}

// NewUserRepository crée une instance de UserRepository.
func NewUserRepository(coll *mongo.Collection) *UserRepository {
	handler := slog.NewTextHandler(os.Stdout, nil)
	return &UserRepository{
		coll:   coll,
		logger: slog.New(handler),
	}
}

func (r *UserRepository) Create(ctx context.Context, u user.User) error {
	return nil
}

func (r *UserRepository) Update(ctx context.Context, u user.User) error {
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (user.User, error) {
	// Implémente la récupération de l'utilisateur par son ID dans MongoDB.
	return user.User{}, errors.New("not implemented")
}

func (r *UserRepository) List(ctx context.Context, filter *user.Filter) ([]user.User, error) {
	// Implémente la récupération d'une liste d'utilisateurs en fonction du filtre.
	return nil, nil
}
