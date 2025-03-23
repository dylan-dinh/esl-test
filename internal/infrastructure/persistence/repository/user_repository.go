package repository

import (
	"context"
	"errors"
	"github.com/dylan-dinh/esl-test/internal/domain/user"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"log/slog"
	"os"
)

const collectionName = "users"

// UserRepository concrete implementation of user.Repository
type UserRepository struct {
	coll   *mongo.Collection
	logger *slog.Logger
}

// NewUserRepository crée une instance de UserRepository.
func NewUserRepository(coll *mongo.Client, dbName string) *UserRepository {
	handler := slog.NewTextHandler(os.Stdout, nil)
	return &UserRepository{
		coll:   coll.Database(dbName).Collection(collectionName),
		logger: slog.New(handler),
	}
}

func (r *UserRepository) Create(ctx context.Context, u user.User) error {
	one, err := r.coll.InsertOne(ctx, u)
	if err != nil {
		return err
	}

	r.logger.Info("insert user with id : %s", one.InsertedID)

	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	filter := bson.D{{"id", u.ID}}
	update := bson.D{{
		"$set", bson.D{
			{"firstname", u.FirstName},
			{"lastname", u.LastName},
			{"nickname", u.Nickname},
			{"email", u.Email},
			{"country", u.Country},
			{"password", u.Password},
			{"updated_at", u.UpdatedAt},
		},
	}}
	res := r.coll.FindOneAndUpdate(ctx, filter, update)
	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			r.logger.Error("user %s not found", u.ID)
			return res.Err()
		}
	}
	r.logger.Info("user %s modified successfully", u.ID)
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
