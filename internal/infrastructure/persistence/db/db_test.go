//go:build integration
// +build integration

package db

import (
	"context"
	"github.com/dylan-dinh/esl-test/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"testing"
	"time"
)

func TestNewDbIntegration(t *testing.T) {
	type testCase struct {
		name          string
		cfg           config.Config
		shouldSucceed bool
	}

	testCases := []testCase{
		{
			name: "Success - valid configuration",
			cfg: config.Config{
				GrpcPort: "50051",
				DbHost:   "localhost",
				DbPort:   "27017",
				DbName:   "testdb",
			},
			shouldSucceed: true,
		},
		{
			name: "Failure - invalid DB port",
			cfg: config.Config{
				GrpcPort: "50051",
				DbHost:   "localhost",
				DbPort:   "12345", // assuming nothing is listening here
				DbName:   "testdb",
			},
			shouldSucceed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dbInstance, err := NewDb(tc.cfg)
			if tc.shouldSucceed {
				require.NoError(t, err, "expected connection to succeed")
				assert.NotNil(t, dbInstance.DB, "mongo client should not be nil")

				// We still ping in the test even though we ping in the NewDB func
				// This is to confirm that the test is ok
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				err = dbInstance.DB.Ping(ctx, readpref.Primary())
				assert.NoError(t, err, "expected ping to succeed")

				err = dbInstance.DB.Disconnect(context.Background())
				require.NoError(t, err, "expected disconnect to succeed")
			} else {
				assert.Error(t, err, "expected error due to invalid configuration")
			}
		})
	}
}
