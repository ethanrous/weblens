package service_test

import (
	"context"
	"sync"
	"testing"

	"github.com/ethrousseau/weblens/models"
	. "github.com/ethrousseau/weblens/service"
	"github.com/ethrousseau/weblens/service/mock"
	"github.com/stretchr/testify/assert"
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
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	userService, err := NewUserService(col)
	if err != nil {
		panic(err)
	}

	// test user 1, do not auto activate
	testUser1, err := models.NewUser(testUser1Name, testUser1Pass, false, false)
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
	testUser2, err := models.NewUser(testUser2Name, testUser1Pass, false, true)
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

	err = col.Drop(context.Background())
	if err != nil {
		panic(err)
	}

	// Test mongo failures
	failingMongo := &mock.MockFailMongoCol{
		RealCol:    col,
		InsertFail: true,
		FindFail:   false,
		UpdateFail: true,
	}

	failUserService, err := NewUserService(failingMongo)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

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

func TestUserServiceImpl_ActivateUser(t *testing.T) {
	type args struct {
		u *models.User
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {

			},
		)
	}
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

	newUser, err := models.NewUser(testUser1Name, testUser1Pass, false, false)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = userService.Add(newUser)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = userService.Del(testUser1Name)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	noUser := userService.Get(testUser1Name)
	assert.Nil(t, noUser)
}

func TestUserServiceImpl_GenerateToken(t *testing.T) {

	// type args struct {
	// 	user *models.User
	// }
	// tests := []struct {
	// 	name    string
	// 	fields  fields
	// 	args    args
	// 	want    string
	// 	wantErr assert.ErrorAssertionFunc
	// }{
	// 	// TODO: Add test cases.
	// }
	// for _, tt := range tests {
	// 	t.Run(
	// 		tt.name, func(t *testing.T) {
	// 			Test
	// 			if !tt.wantErr(t, err, fmt.Sprintf("GenerateToken(%v)", tt.args.user)) {
	// 				return
	// 			}
	// 			assert.Equalf(t, tt.want, got, "GenerateToken(%v)", tt.args.user)
	// 		},
	// 	)
	// }
}

func TestUserServiceImpl_SearchByUsername(t *testing.T) {

	// type args struct {
	// 	searchString string
	// }
	// tests := []struct {
	// 	name    string
	// 	fields  fields
	// 	args    args
	// 	want    iter.Seq[*models.User]
	// 	wantErr assert.ErrorAssertionFunc
	// }{
	// 	// TODO: Add test cases.
	// }
	// for _, tt := range tests {
	// 	t.Run(
	// 		tt.name, func(t *testing.T) {
	// 			Test
	// 			if !tt.wantErr(t, err, fmt.Sprintf("SearchByUsername(%v)", tt.args.searchString)) {
	// 				return
	// 			}
	// 			assert.Equalf(t, tt.want, got, "SearchByUsername(%v)", tt.args.searchString)
	// 		},
	// 	)
	// }
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

	newUser, err := models.NewUser(testUser1Name, testUser1Pass, false, false)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = userService.Add(newUser)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = userService.SetUserAdmin(newUser, true)
	assert.Error(t, err)
	assert.False(t, newUser.IsAdmin())

	err = userService.ActivateUser(newUser)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

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
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	newUser2, err := models.NewUser(testUser2Name, testUser2Pass, false, true)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, newUser2.IsActive())

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

	newUser, err := models.NewUser(testUser1Name, testUser1Pass, false, true)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = userService.Add(newUser)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = userService.UpdateUserPassword(newUser.Username, testUser1Pass, testUser2Pass, false)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.False(t, newUser.CheckLogin(testUser1Pass))
	assert.True(t, newUser.CheckLogin(testUser2Pass))
}
