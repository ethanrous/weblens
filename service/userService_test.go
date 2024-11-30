package service_test

import (
	"context"
	"slices"
	"sync"
	"testing"

	"github.com/ethanrous/weblens/models"
	. "github.com/ethanrous/weblens/service"
	"github.com/ethanrous/weblens/service/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	testUser1Name = "testUser1"
	testUser2Name = "testUser2"
	testUser1Pass = "testPass1"
	testUser2Pass = "testPass2"
)

func TestUserService(t *testing.T) {
	t.Parallel()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	require.NoError(t, err)

	defer col.Drop(context.Background())

	userService, err := NewUserService(col)
	require.NoError(t, err)

	fs := &mock.MockFileService{}

	// test user 1, do not auto activate
	testUser1, err := models.NewUser(testUser1Name, testUser1Pass, false, false)
	require.NoError(t, err)

	err = fs.CreateUserHome(testUser1)
	require.NoError(t, err)

	err = userService.Add(testUser1)
	require.NoError(t, err)

	serviceUser1 := userService.Get(testUser1Name)
	assert.NotNil(t, serviceUser1)

	assert.False(t, serviceUser1.IsActive())
	err = userService.ActivateUser(testUser1, true)
	require.NoError(t, err)

	assert.True(t, serviceUser1.IsActive())
	assert.Equal(t, 1, userService.Size())

	// test user 2, do auto activate
	testUser2, err := models.NewUser(testUser2Name, testUser1Pass, false, true)
	require.NoError(t, err)

	err = fs.CreateUserHome(testUser2)
	require.NoError(t, err)

	err = userService.Add(testUser2)
	require.NoError(t, err)

	serviceUser2 := userService.Get(testUser2Name)
	assert.NotNil(t, serviceUser2)

	assert.True(t, serviceUser2.IsActive())
	assert.Equal(t, 2, userService.Size())

	err = userService.Del(testUser1Name)
	require.NoError(t, err)

	err = userService.Del(testUser2Name)
	require.NoError(t, err)

	assert.Equal(t, 0, userService.Size())

	err = col.Drop(context.Background())
	require.NoError(t, err)

	// Test mongo failures
	failingMongo := &mock.MockFailMongoCol{
		RealCol:    col,
		InsertFail: true,
		FindFail:   false,
		UpdateFail: true,
	}

	failUserService, err := NewUserService(failingMongo)
	require.NoError(t, err)

	err = failUserService.Add(serviceUser1)
	assert.Error(t, err)
	assert.Nil(t, failUserService.Get(testUser1Name))
}

type fields struct {
	userMap    map[models.Username]*models.User
	userLock   sync.RWMutex
	publicUser *models.User
	rootUser   *models.User
	col        *mongo.Collection
}

func TestUserServiceImpl_Add(t *testing.T) {
	t.Parallel()
	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	userService, err := NewUserService(col)
	if err != nil {
		panic(err)
	}

	_, err = models.NewUser(testUser1Name, "", false, false)
	assert.Error(t, err)
	_, err = models.NewUser("", testUser1Pass, false, false)
	assert.Error(t, err)

	badUser := &models.User{
		Username: "",
		Password: "",
	}

	err = userService.Add(badUser)
	assert.Error(t, err)
}

func TestUserServiceImpl_Del(t *testing.T) {
	t.Parallel()
	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	userService, err := NewUserService(col)
	if err != nil {
		panic(err)
	}

	fs := &mock.MockFileService{}

	newUser, err := models.NewUser(testUser1Name, testUser1Pass, false, false)
	require.NoError(t, err)

	err = fs.CreateUserHome(newUser)
	require.NoError(t, err)

	err = userService.Add(newUser)
	require.NoError(t, err)

	err = userService.Del(testUser1Name)
	require.NoError(t, err)

	noUser := userService.Get(testUser1Name)
	assert.Nil(t, noUser)
}

func TestUserServiceImpl_SearchByUsername(t *testing.T) {
	t.Parallel()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	userService, err := NewUserService(col)
	if err != nil {
		panic(err)
	}

	fs := &mock.MockFileService{}

	stanUser, err := models.NewUser("stan", testUser1Pass, false, true)
	require.NoError(t, err)

	err = fs.CreateUserHome(stanUser)
	require.NoError(t, err)

	err = userService.Add(stanUser)
	require.NoError(t, err)

	blanUser, err := models.NewUser("blan", testUser1Pass, false, true)
	require.NoError(t, err)

	err = fs.CreateUserHome(blanUser)
	require.NoError(t, err)

	err = userService.Add(blanUser)
	require.NoError(t, err)

	userIter, err := userService.SearchByUsername("st")
	require.NoError(t, err)

	userResults := slices.Collect(userIter)
	assert.Equal(t, 1, len(userResults))
	if len(userResults) != 0 {
		assert.Equal(t, "stan", userResults[0].Username)
	}

	userIter2, err := userService.SearchByUsername("an")
	require.NoError(t, err)

	userResults2 := slices.Collect(userIter2)
	assert.Equal(t, 2, len(userResults2))
}

func TestUserServiceImpl_SetUserAdmin(t *testing.T) {
	t.Parallel()
	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	userService, err := NewUserService(col)
	if err != nil {
		panic(err)
	}

	fs := &mock.MockFileService{}

	newUser, err := models.NewUser(testUser1Name, testUser1Pass, false, false)
	require.NoError(t, err)

	err = fs.CreateUserHome(newUser)
	require.NoError(t, err)

	err = userService.Add(newUser)
	require.NoError(t, err)

	err = userService.SetUserAdmin(newUser, true)
	assert.Error(t, err)
	assert.False(t, newUser.IsAdmin())

	err = userService.ActivateUser(newUser, true)
	require.NoError(t, err)

	err = userService.SetUserAdmin(newUser, true)
	assert.NoError(t, err)
	assert.True(t, newUser.IsAdmin())

	// Test mongo failures
	failingMongo := &mock.MockFailMongoCol{
		RealCol:    col,
		InsertFail: false,
		FindFail:   false,
		UpdateFail: true,
	}

	failUserService, err := NewUserService(failingMongo)
	require.NoError(t, err)

	newUser2, err := models.NewUser(testUser2Name, testUser2Pass, false, true)
	require.NoError(t, err)
	assert.True(t, newUser2.IsActive())

	err = fs.CreateUserHome(newUser2)
	require.NoError(t, err)

	err = failUserService.Add(newUser2)
	assert.NoError(t, err)
	assert.NotNil(t, failUserService.Get(testUser1Name))
	assert.False(t, newUser2.IsAdmin())

	err = failUserService.SetUserAdmin(newUser2, true)
	assert.Error(t, err)
	assert.False(t, newUser2.IsAdmin())
}

func TestUserServiceImpl_UpdateUserPassword(t *testing.T) {
	t.Parallel()
	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	userService, err := NewUserService(col)
	if err != nil {
		panic(err)
	}

	fs := &mock.MockFileService{}

	newUser, err := models.NewUser(testUser1Name, testUser1Pass, false, true)
	require.NoError(t, err)

	err = fs.CreateUserHome(newUser)
	require.NoError(t, err)

	err = userService.Add(newUser)
	require.NoError(t, err)

	err = userService.UpdateUserPassword(newUser.Username, testUser1Pass, testUser2Pass, false)
	require.NoError(t, err)

	assert.False(t, newUser.CheckLogin(testUser1Pass))
	assert.True(t, newUser.CheckLogin(testUser2Pass))
}
