package repository

import (
	"context"
	"errors"
	"github.com/dylan-dinh/esl-test/internal/domain/user"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
func NewUserRepository(conn *mongo.Client, dbName string) *UserRepository {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	coll := conn.Database(dbName).Collection(collectionName)

	constraint := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	if _, err := coll.Indexes().CreateOne(context.Background(), constraint); err != nil {
		var cmdErr mongo.CommandError
		if errors.As(err, &cmdErr) && cmdErr.Code == 85 {
			// index already exists, safe to ignore
		} else {
			logger.Error("fail to create unicity on email")
		}
	}
	logger.Info("unicity on users.email created")

	return &UserRepository{
		coll:   coll,
		logger: logger,
	}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	_, err := r.coll.InsertOne(ctx, &u)
	if err != nil {
		return err
	}

	r.logger.Info("insert user with id", "id", u.ID)

	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	filter := bson.D{{"id", u.ID}}
	update := bson.D{{
		"$set", bson.D{
			{"first_name", u.FirstName},
			{"last_name", u.LastName},
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
			r.logger.Error("user not found", "id", "id", u.ID, u.ID)
			return res.Err()
		}
	}
	r.logger.Info("user modified successfully", "id", u.ID)
	return nil
}

func (r *UserRepository) DeleteByID(ctx context.Context, id string) error {
	filter := bson.D{{"id", id}}

	res, err := r.coll.DeleteOne(ctx, filter)
	if err != nil {
		r.logger.Error("error deleting user", "id", id, "error", err)
		return err
	}

	if res.DeletedCount == 0 {
		r.logger.Error("user not found", "id", id)
		return mongo.ErrNoDocuments
	}

	r.logger.Info("user deleted", "id", id)
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (user.User, error) {
	filter := bson.D{{"id", id}}

	opts := options.FindOne().SetProjection(bson.D{{"password", 0}})

	res := r.coll.FindOne(ctx, filter, opts)
	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			r.logger.Error("user not found", "id", id)
			return user.User{}, res.Err()
		}
	}

	var getUser user.User
	err := res.Decode(&getUser)
	if err != nil {
		r.logger.Error("couldn't decode result from mongo")
		return user.User{}, err
	}

	return getUser, nil

}

func (r *UserRepository) List(ctx context.Context, filter *user.UserFilter) ([]user.User, int64, error) {
	query := bson.D{}
	if filter.FirstName != "" {
		query = append(query, bson.E{Key: "first_name", Value: filter.FirstName})
	}
	if filter.LastName != "" {
		query = append(query, bson.E{Key: "last_name", Value: filter.LastName})
	}
	if filter.Country != "" {
		query = append(query, bson.E{Key: "country", Value: filter.Country})
	}

	total, err := r.coll.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((filter.Page - 1) * filter.PageSize)
	limit := int64(filter.PageSize)
	opts := options.Find().SetSkip(skip).SetLimit(limit)

	cursor, err := r.coll.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	var users []user.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	filter := bson.D{{Key: "email", Value: email}}
	err := r.coll.FindOne(ctx, filter).Err()

	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}
