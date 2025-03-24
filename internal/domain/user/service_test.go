package user

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

type fakeRepo struct {
	exists bool
	err    error
}

func (f *fakeRepo) Create(ctx context.Context, u *User) error            { return nil }
func (f *fakeRepo) Update(ctx context.Context, u *User) error            { return nil }
func (f *fakeRepo) DeleteByID(ctx context.Context, id string) error      { return nil }
func (f *fakeRepo) GetByID(ctx context.Context, id string) (User, error) { return User{}, nil }
func (f *fakeRepo) List(ctx context.Context, filter *UserFilter) ([]User, int64, error) {
	return []User{}, 0, nil
}
func (f *fakeRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return f.exists, f.err
}

func TestCreateUserValidation(t *testing.T) {
	cases := []struct {
		name    string
		input   *User
		repo    *fakeRepo
		wantErr error
	}{
		{"missing email",
			&User{
				FirstName: "foo",
				LastName:  "bar",
				Password:  "password",
			},
			&fakeRepo{},
			ErrMissingEmailPassword,
		},
		{"duplicate email",
			&User{
				Email:     "x@example.com",
				Password:  "password",
				FirstName: "foo",
				LastName:  "bar",
			},
			&fakeRepo{exists: true},
			ErrEmailExists,
		},
		{"success",
			&User{
				Email:     "x@example.com",
				Password:  "password",
				FirstName: "foo",
				LastName:  "bar",
			},
			&fakeRepo{exists: false},
			nil,
		},
		{"missing password",
			&User{
				Email:     "x@example.com",
				FirstName: "foo",
				LastName:  "bar",
			},
			&fakeRepo{exists: true},
			ErrMissingEmailPassword,
		},
		{"missing first name",
			&User{
				Email:    "x@example.com",
				LastName: "bar",
				Password: "password",
			},
			&fakeRepo{exists: true},
			ErrMissingName,
		},
		{"missing first name",
			&User{
				Email:     "x@example.com",
				FirstName: "bar",
				Password:  "password",
			},
			&fakeRepo{exists: true},
			ErrMissingName,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewUserService(tc.repo)
			err := svc.CreateUser(context.Background(), tc.input)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tc.input.ID)
			}
		})
	}
}
