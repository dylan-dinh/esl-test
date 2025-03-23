//go:build integration
// +build integration

package user

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"

	"github.com/dylan-dinh/esl-test/internal/config"
	"github.com/dylan-dinh/esl-test/internal/domain/user"
	"github.com/dylan-dinh/esl-test/internal/infrastructure/persistence/db"
	"github.com/dylan-dinh/esl-test/internal/infrastructure/persistence/repository"
	"github.com/stretchr/testify/assert"
)

// setupIntegrationTest creates a temporary directory, writes a .env file in it
// creates the DB connection and returns the user service and a cleanup function.
func setupIntegrationTest(t *testing.T) (user.Service, func()) {
	t.Helper()

	// Write a temporary .env file.
	envContent := `GRPC_PORT=50051
DB_HOST=localhost
DB_PORT=27017
DB_NAME=testdb
`
	err := os.WriteFile(".env", []byte(envContent), 0644)
	require.NoError(t, err)

	conf, err := config.GetConfig()
	require.NoError(t, err)

	newDb, err := db.NewDb(conf)
	require.NoError(t, err)

	userRepo := repository.NewUserRepository(newDb.DB, conf.DbName)
	userSvc := user.NewUserService(userRepo)

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = newDb.DB.Disconnect(ctx)
		_ = os.Remove(".env")
	}

	return userSvc, cleanup
}

func TestCreateUserIntegration(t *testing.T) {
	userSvc, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	testUser := &user.User{
		FirstName: "Test",
		LastName:  "User",
		Nickname:  "test_user",
		Email:     "testuser@faceit.com",
		Country:   "FR",
		Password:  "password",
	}

	err := userSvc.CreateUser(ctx, testUser)
	assert.NoError(t, err, "CreateUser should not return an error")
	assert.NotEmpty(t, testUser.ID, "User ID should be generated")
}

func TestUpdateUserIntegration(t *testing.T) {
	userSvc, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	testUser := &user.User{
		FirstName: "Test",
		LastName:  "User",
		Nickname:  "test_user",
		Email:     "testuser@faceit.com",
		Country:   "FR",
		Password:  "password",
	}
	err := userSvc.CreateUser(ctx, testUser)
	assert.NoError(t, err, "CreateUser should not return an error")
	assert.NotEmpty(t, testUser.ID, "User ID should be generated")

	// Update some fields
	testUser.FirstName = "New First Name"
	testUser.LastName = "New Last Name"
	testUser.Nickname = "New Nickname"
	testUser.Email = "updated@faceit.com"
	testUser.Country = "UK"
	testUser.Password = "newplainpassword"
	testUser.UpdatedAt = time.Now()

	err = userSvc.UpdateUser(ctx, testUser)
	assert.NoError(t, err, "UpdateUser should not return an error")

	// Retrieve the user via the GetUser method to verify the update
	updatedUser, err := userSvc.GetUser(ctx, testUser.ID)
	assert.NoError(t, err, "GetUser should not return an error")
	assert.Equal(t, "New First Name", updatedUser.FirstName, "FirstName should be updated")
	assert.Equal(t, "New Last Name", updatedUser.LastName, "LastName should be updated")
	assert.Equal(t, "New Nickname", updatedUser.Nickname, "Nickname should be updated")
	assert.Equal(t, "updated@faceit.com", updatedUser.Email, "Email should be updated")
	assert.Equal(t, "UK", updatedUser.Country, "Country should be updated")
}

func TestDeleteUserIntegration(t *testing.T) {
	userSvc, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	testUser := &user.User{
		FirstName: "Test",
		LastName:  "User",
		Nickname:  "test_user",
		Email:     "testuser@faceit.com",
		Country:   "FR",
		Password:  "password",
	}

	err := userSvc.CreateUser(ctx, testUser)
	require.NoError(t, err, "CreateUser should succeed")
	require.NotEmpty(t, testUser.ID, "User ID should be generated")

	err = userSvc.DeleteUser(ctx, testUser.ID)
	require.NoError(t, err, "DeleteUser should succeed")

	_, err = userSvc.GetUser(ctx, testUser.ID)
	assert.Error(t, err, "Expected error when retrieving a deleted user")
}

func TestGetUserIntegration(t *testing.T) {
	userSvc, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	testUser := &user.User{
		FirstName: "Test",
		LastName:  "User",
		Nickname:  "test_user",
		Email:     "testuser@faceit.com",
		Country:   "FR",
		Password:  "password",
	}

	err := userSvc.CreateUser(ctx, testUser)
	require.NoError(t, err, "CreateUser should succeed")
	require.NotEmpty(t, testUser.ID, "User ID should be generated")

	returnedUser, err := userSvc.GetUser(ctx, testUser.ID)
	assert.NoError(t, err, "Expected no error when retrieving a user")

	assert.Equal(t, testUser.FirstName, returnedUser.FirstName)
	assert.Equal(t, testUser.LastName, returnedUser.LastName)
	assert.Equal(t, testUser.Nickname, returnedUser.Nickname)
	assert.Equal(t, testUser.Email, returnedUser.Email)
	assert.Equal(t, testUser.Country, returnedUser.Country)
}
