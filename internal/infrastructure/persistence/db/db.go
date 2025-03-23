package db

import (
	"context"
	"fmt"
	"github.com/dylan-dinh/esl-test/internal/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"time"
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

	// Create a context with a timeout for the ping.
	// After 3 seconds, we consider that mongoDB is down
	// This will affect by 3 seconds the failure test
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return DB{}, err
	}

	return DB{
		DB: client,
	}, nil
}
