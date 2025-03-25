package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	type testCase struct {
		name          string
		envContent    string
		shouldSucceed bool
		expected      Config
		missingKey    string
	}

	testCases := []testCase{
		{
			name: "Success - all env provided",
			envContent: `GRPC_PORT=50051
DB_HOST=localhost
DB_PORT=27017
DB_NAME=testdb
RABBIT_HOST=rabbitmq
RABBIT_PORT=5672
`,
			shouldSucceed: true,
			expected: Config{
				GrpcPort: "50051",
				DbHost:   "localhost",
				DbPort:   "27017",
				DbName:   "testdb",
			},
		},
		{
			name: "Failure - missing GRPC_PORT",
			envContent: `DB_HOST=localhost
DB_PORT=27017
DB_NAME=testdb
`,
			shouldSucceed: false,
			missingKey:    "GRPC_PORT",
		},
		{
			name: "Failure - missing DB_HOST",
			envContent: `GRPC_PORT=50051
DB_PORT=27017
DB_NAME=testdb
`,
			shouldSucceed: false,
			missingKey:    "DB_HOST",
		},
		{
			name: "Failure - missing DB_PORT",
			envContent: `GRPC_PORT=50051
DB_HOST=localhost
DB_NAME=testdb
`,
			shouldSucceed: false,
			missingKey:    "DB_PORT",
		},
		{
			name: "Failure - missing DB_NAME",
			envContent: `GRPC_PORT=50051
DB_HOST=localhost
DB_PORT=27017
`,
			shouldSucceed: false,
			missingKey:    "DB_NAME",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear environment variables before running each test case
			os.Clearenv()

			// Write temporary .env file
			err := os.WriteFile(".env", []byte(tc.envContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create temporary .env file: %v", err)
			}
			// Ensure .env is removed after the test
			defer os.Remove(".env")

			conf, err := GetConfig()
			if tc.shouldSucceed {
				assert.NoError(t, err, "expected no error when all env vars are provided")
				assert.Equal(t, tc.expected.GrpcPort, conf.GrpcPort, "expected GRPC_PORT to match")
				assert.Equal(t, tc.expected.DbHost, conf.DbHost, "expected DB_HOST to match")
				assert.Equal(t, tc.expected.DbPort, conf.DbPort, "expected DB_PORT to match")
				assert.Equal(t, tc.expected.DbName, conf.DbName, "expected DB_NAME to match")
			} else {
				assert.Error(t, err, "expected error due to missing %s", tc.missingKey)
				assert.Contains(t, err.Error(), tc.missingKey, "error message should contain missing key")
			}
		})
	}
}
