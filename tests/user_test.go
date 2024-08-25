package tests

import (
	"testing"

	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/models/service"
	"github.com/stretchr/testify/assert"
)

const (
	testUser1Name = "testUser1"
	testUser2Name = "testUser2"
	testUserPass  = "testPass"
)

func TestUserService(t *testing.T) {
	// store := newMockUserStore()
	userService := service.NewUserService(nil)
	// err := userService.Init(store)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// test user 1, do not auto activate
	testUser1, err := models.NewUser(testUser1Name, testUserPass, false, false)
	if err != nil {
		t.Fatal(err)
	}

	err = userService.Add(testUser1)
	if err != nil {
		t.Fatal(err)
	}

	serviceUser1 := userService.Get(testUser1Name)
	assert.NotNil(t, serviceUser1)

	assert.False(t, serviceUser1.IsActive())
	err = userService.ActivateUser(testUser1)

	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, serviceUser1.IsActive())
	assert.Equal(t, 1, userService.Size())

	// test user 2, do auto activate
	testUser2, err := models.NewUser(testUser2Name, testUserPass, false, true)
	if err != nil {
		t.Fatal(err)
	}

	err = userService.Add(testUser2)
	if err != nil {
		t.Fatal(err)
	}

	serviceUser2 := userService.Get(testUser2Name)
	assert.NotNil(t, serviceUser2)

	assert.True(t, serviceUser2.IsActive())
	assert.Equal(t, 2, userService.Size())

	err = userService.Del(testUser1Name)
	if err != nil {
		t.Fatal(err)
	}

	err = userService.Del(testUser2Name)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 0, userService.Size())
}

type mockUserStore struct {
}

func (us *mockUserStore) GetAllUsers() ([]*models.User, error) {
	return []*models.User{}, nil
}

func (us *mockUserStore) UpdatePasswordByUsername(username models.Username, newPasswordHash string) error {
	return nil
}

func (us *mockUserStore) SetAdminByUsername(username models.Username, b bool) error {
	return nil
}

func (us *mockUserStore) CreateUser(user *models.User) error {
	return nil
}

func (us *mockUserStore) ActivateUser(username models.Username) error {
	return nil
}

func (us *mockUserStore) AddTokenToUser(username models.Username, token string) error {
	return nil
}

func (us *mockUserStore) SearchUsers(search string) ([]models.Username, error) {
	// TODO implement me
	panic("implement me")
}

func (us *mockUserStore) DeleteUserByUsername(username models.Username) error {
	return nil
}

func (us *mockUserStore) DeleteAllUsers() error {
	return nil
}

// func newMockUserStore() *weblens.UserStore {
// 	return &mockUserStore{}
// }
