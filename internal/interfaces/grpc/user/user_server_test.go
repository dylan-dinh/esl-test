package user

import (
	"context"
	"encoding/json"
	"github.com/dylan-dinh/esl-test/internal/domain/user"
	"github.com/dylan-dinh/esl-test/internal/interfaces/notifier"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/dylan-dinh/esl-test/internal/config"
	"github.com/dylan-dinh/esl-test/internal/infrastructure/persistence/db"
	"github.com/dylan-dinh/esl-test/internal/infrastructure/persistence/repository"
	"github.com/stretchr/testify/assert"
)

func declareAndBindQueue(t *testing.T, mq *user.RabbitMQ, routingKey string) string {
	q, err := mq.Ch.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)
	err = mq.Ch.QueueBind(q.Name, routingKey, "user.events", false, nil)
	require.NoError(t, err)

	return q.Name
}

func assertPublished[T any](t *testing.T, mq *user.RabbitMQ, queueName string, assertFn func(t *testing.T, actual T)) {
	msgs, err := mq.Ch.Consume(queueName, "", true, true, false, false, nil)
	assert.NoError(t, err)
	select {
	case msg := <-msgs:
		var actual T
		require.NoError(t, json.Unmarshal(msg.Body, &actual))
		// I did this because I had mismatch on timestamp
		// when comparing whole User object
		if assertFn != nil {
			assertFn(t, actual)
		} else {
			t.Log("nothing to compare")
		}
	case <-time.After(4 * time.Second):
		t.Fatalf("expected event but none arrived")
	}
}

// setupIntegrationTest creates a temporary directory, writes a .env file in it
// creates the DB connection and returns the user service and a cleanup function.
func setupIntegrationTest(t *testing.T) (user.Service, func(), *user.RabbitMQ) {
	t.Helper()

	conf, err := config.GetConfig()
	require.NoError(t, err)

	newDb, err := db.NewDb(conf)
	require.NoError(t, err)

	rabbitConn, err := notifier.NewRabbitMQConn(conf)
	require.NoError(t, err)

	mq, err := user.NewRabbitMQ(rabbitConn)
	require.NoError(t, err)

	userRepo, err := repository.NewUserRepository(newDb.DB, conf.DbName)
	require.NoError(t, err)
	userSvc := user.NewUserService(userRepo, mq)

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = newDb.DB.Database(conf.DbName).Collection("users").Drop(ctx)
		_ = newDb.DB.Disconnect(ctx)
		_ = rabbitConn.Close()
	}

	return userSvc, cleanup, mq
}

func TestCreateUserIntegration(t *testing.T) {
	userSvc, cleanup, mq := setupIntegrationTest(t)
	defer cleanup()

	queueName := declareAndBindQueue(t, mq, user.UserCreatedRoutingKey)

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
	assertPublished[user.User](t, mq, queueName, func(t *testing.T, actual user.User) {
		assert.Equal(t, testUser.ID, actual.ID)
	})
	assert.NoError(t, err, "CreateUser should not return an error")
	assert.NotEmpty(t, testUser.ID, "User ID should be generated")
}

func TestUpdateUserIntegration(t *testing.T) {
	userSvc, cleanup, mq := setupIntegrationTest(t)
	defer cleanup()

	queueName := declareAndBindQueue(t, mq, user.UserUpdatedRoutingKey)

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

	assertPublished[user.User](t, mq, queueName, func(t *testing.T, actual user.User) {
		assert.Equal(t, testUser.ID, actual.ID)
	})

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
	userSvc, cleanup, mq := setupIntegrationTest(t)
	defer cleanup()

	queueName := declareAndBindQueue(t, mq, user.UserDeletedRoutingKey)

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

	assertPublished[string](t, mq, queueName, func(t *testing.T, actual string) {
		assert.Equal(t, testUser.ID, actual)
	})

	_, err = userSvc.GetUser(ctx, testUser.ID)
	assert.Error(t, err, "Expected error when retrieving a deleted user")
}

func TestGetUserIntegration(t *testing.T) {
	userSvc, cleanup, _ := setupIntegrationTest(t)
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

func TestListUsersByFirstNameIntegration(t *testing.T) {
	userSvc, cleanup, _ := setupIntegrationTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create test data.
	testUsers := []user.User{
		{FirstName: "Dylan", LastName: "Dinh", Country: "UK", Nickname: "DD1", Email: "Dylan1@faceit.com", Password: "password"},
		{FirstName: "Dylan", LastName: "Brown", Country: "UK", Nickname: "DB1", Email: "Dylan2@faceit.com", Password: "password"},
		{FirstName: "Foo", LastName: "Jones", Country: "US", Nickname: "FJ1", Email: "Foo1@faceit.com", Password: "password"},
		{FirstName: "Marko", LastName: "White", Country: "UK", Nickname: "MW1", Email: "Marko@faceit.com", Password: "password"},
		{FirstName: "Matteo", LastName: "Black", Country: "US", Nickname: "MB1", Email: "Matteo@faceit.com", Password: "password"},
		{FirstName: "Dylan", LastName: "Cooper", Country: "UK", Nickname: "DC1", Email: "Dylan3@faceit.com", Password: "password"},
		{FirstName: "Etienne", LastName: "Davis", Country: "US", Nickname: "ED1", Email: "Etienne@faceit.com", Password: "password"},
		{FirstName: "Thomas", LastName: "Miller", Country: "UK", Nickname: "TM1", Email: "Thomas@faceit.com", Password: "password"},
	}

	// Insert all test users.
	for i := range testUsers {
		err := userSvc.CreateUser(ctx, &testUsers[i])
		require.NoError(t, err, "CreateUser should succeed")
		require.NotEmpty(t, testUsers[i].ID, "User ID should be generated")
	}

	// Define test cases.
	type listTestCase struct {
		name          string
		firstName     string
		lastName      string
		country       string
		page          int32
		pageSize      int32
		expectedTotal int64 // Total count of matching users
		expectedSlice int   // Number of users returned in the page
	}

	testCases := []listTestCase{
		{
			name:          "Filter by first name 'Dylan'",
			firstName:     "Dylan",
			lastName:      "",
			country:       "",
			page:          1,
			pageSize:      10,
			expectedTotal: 3, // Three users with first name "Dylan"
			expectedSlice: 3,
		},
		{
			name:          "Filter by last name 'Dinh'",
			firstName:     "",
			lastName:      "Dinh",
			country:       "",
			page:          1,
			pageSize:      10,
			expectedTotal: 1, // One user with last name "Dinh"
			expectedSlice: 1,
		},
		{
			name:          "Filter by country 'UK'",
			firstName:     "",
			lastName:      "",
			country:       "UK",
			page:          1,
			pageSize:      10,
			expectedTotal: 5, // Users: Dylan Dinh, Dylan Brown, Charlie White, Dylan Cooper, Frank Miller
			expectedSlice: 5,
		},
		{
			name:          "Pagination: page 1 with 10 per page",
			firstName:     "",
			lastName:      "",
			country:       "", // No filtering, list all users
			page:          1,
			pageSize:      10,
			expectedTotal: 8, // Total 8 users inserted
			expectedSlice: 8,
		},
		{
			name:          "Pagination: page 2 with 2 per page",
			firstName:     "",
			lastName:      "",
			country:       "", // No filtering, list all users
			page:          2,
			pageSize:      2,
			expectedTotal: 8, // Total 8 users inserted
			expectedSlice: 2,
		},
	}

	// Iterate over each test case.
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filter := &user.UserFilter{
				FirstName: tc.firstName,
				LastName:  tc.lastName,
				Country:   tc.country,
				Page:      tc.page,
				PageSize:  tc.pageSize,
			}
			listedUsers, total, err := userSvc.ListUsers(ctx, filter)
			require.NoError(t, err, "ListUsers should succeed")
			assert.Equal(t, tc.expectedTotal, total, "expected total count")
			assert.Equal(t, tc.expectedSlice, len(listedUsers), "expected slice length")
		})
	}
}
