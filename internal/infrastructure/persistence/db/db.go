package db

import (
	"fmt"
	"github.com/dylan-dinh/esl-test/internal/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DB struct {
	DB *mongo.Client
}

// NewDb handle connection to database and to collection directly since
// we only manage user entity
func NewDb(config config.Config) (DB, error) {
	client, err := mongo.Connect(options.Client().
		ApplyURI(fmt.Sprintf("mongodb://%s:%s", config.DbHost, config.DbPort)))
	if err != nil {
		return DB{}, err
	}

	return DB{
		DB: client,
	}, nil
}
