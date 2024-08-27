package service

import (
	"context"
	"fmt"
	"iter"
	"sync"
	"testing"

	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	testUser1Name = "testUser1"
	testUser2Name = "testUser2"
	testUserPass  = "testPass"
)

func TestUserService(t *testing.T) {
	db := database.ConnectToMongo("mongodb://localhost:27017", "weblens-test")
	// db := database.ConnectToMongo(internal.GetMongoURI(), "weblens-test")
	err := db.Collection("users").Drop(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	userService := NewUserService(db.Collection("users"))
	err = userService.Init()
	if err != nil {
		panic(err)
	}

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
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				us := &UserServiceImpl{
					userMap:    tt.fields.userMap,
					userLock:   tt.fields.userLock,
					publicUser: tt.fields.publicUser,
					rootUser:   tt.fields.rootUser,
					col:        tt.fields.col,
				}
				tt.wantErr(t, us.ActivateUser(tt.args.u), fmt.Sprintf("ActivateUser(%v)", tt.args.u))
			},
		)
	}
}

func TestUserServiceImpl_Add(t *testing.T) {
	type args struct {
		user *models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				us := &UserServiceImpl{
					userMap:    tt.fields.userMap,
					userLock:   tt.fields.userLock,
					publicUser: tt.fields.publicUser,
					rootUser:   tt.fields.rootUser,
					col:        tt.fields.col,
				}
				tt.wantErr(t, us.Add(tt.args.user), fmt.Sprintf("Add(%v)", tt.args.user))
			},
		)
	}
}

func TestUserServiceImpl_Del(t *testing.T) {
	type args struct {
		un models.Username
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				us := &UserServiceImpl{
					userMap:    tt.fields.userMap,
					userLock:   tt.fields.userLock,
					publicUser: tt.fields.publicUser,
					rootUser:   tt.fields.rootUser,
					col:        tt.fields.col,
				}
				tt.wantErr(t, us.Del(tt.args.un), fmt.Sprintf("Del(%v)", tt.args.un))
			},
		)
	}
}

func TestUserServiceImpl_GenerateToken(t *testing.T) {

	type args struct {
		user *models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				us := &UserServiceImpl{
					userMap:    tt.fields.userMap,
					userLock:   tt.fields.userLock,
					publicUser: tt.fields.publicUser,
					rootUser:   tt.fields.rootUser,
					col:        tt.fields.col,
				}
				got, err := us.GenerateToken(tt.args.user)
				if !tt.wantErr(t, err, fmt.Sprintf("GenerateToken(%v)", tt.args.user)) {
					return
				}
				assert.Equalf(t, tt.want, got, "GenerateToken(%v)", tt.args.user)
			},
		)
	}
}

func TestUserServiceImpl_Get(t *testing.T) {

	type args struct {
		username models.Username
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *models.User
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				us := &UserServiceImpl{
					userMap:    tt.fields.userMap,
					userLock:   tt.fields.userLock,
					publicUser: tt.fields.publicUser,
					rootUser:   tt.fields.rootUser,
					col:        tt.fields.col,
				}
				assert.Equalf(t, tt.want, us.Get(tt.args.username), "Get(%v)", tt.args.username)
			},
		)
	}
}

func TestUserServiceImpl_GetAll(t *testing.T) {

	tests := []struct {
		name   string
		fields fields
		want   iter.Seq[*models.User]
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				us := &UserServiceImpl{
					userMap:    tt.fields.userMap,
					userLock:   tt.fields.userLock,
					publicUser: tt.fields.publicUser,
					rootUser:   tt.fields.rootUser,
					col:        tt.fields.col,
				}
				assert.Equalf(t, tt.want, us.GetAll(), "GetAll()")
			},
		)
	}
}

func TestUserServiceImpl_GetPublicUser(t *testing.T) {

	tests := []struct {
		name   string
		fields fields
		want   *models.User
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				us := &UserServiceImpl{
					userMap:    tt.fields.userMap,
					userLock:   tt.fields.userLock,
					publicUser: tt.fields.publicUser,
					rootUser:   tt.fields.rootUser,
					col:        tt.fields.col,
				}
				assert.Equalf(t, tt.want, us.GetPublicUser(), "GetPublicUser()")
			},
		)
	}
}

func TestUserServiceImpl_GetRootUser(t *testing.T) {

	tests := []struct {
		name   string
		fields fields
		want   *models.User
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				us := &UserServiceImpl{
					userMap:    tt.fields.userMap,
					userLock:   tt.fields.userLock,
					publicUser: tt.fields.publicUser,
					rootUser:   tt.fields.rootUser,
					col:        tt.fields.col,
				}
				assert.Equalf(t, tt.want, us.GetRootUser(), "GetRootUser()")
			},
		)
	}
}

func TestUserServiceImpl_Init(t *testing.T) {

	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				us := &UserServiceImpl{
					userMap:    tt.fields.userMap,
					userLock:   tt.fields.userLock,
					publicUser: tt.fields.publicUser,
					rootUser:   tt.fields.rootUser,
					col:        tt.fields.col,
				}
				tt.wantErr(t, us.Init(), fmt.Sprintf("Init()"))
			},
		)
	}
}

func TestUserServiceImpl_SearchByUsername(t *testing.T) {

	type args struct {
		searchString string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    iter.Seq[*models.User]
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				us := &UserServiceImpl{
					userMap:    tt.fields.userMap,
					userLock:   tt.fields.userLock,
					publicUser: tt.fields.publicUser,
					rootUser:   tt.fields.rootUser,
					col:        tt.fields.col,
				}
				got, err := us.SearchByUsername(tt.args.searchString)
				if !tt.wantErr(t, err, fmt.Sprintf("SearchByUsername(%v)", tt.args.searchString)) {
					return
				}
				assert.Equalf(t, tt.want, got, "SearchByUsername(%v)", tt.args.searchString)
			},
		)
	}
}

func TestUserServiceImpl_SetUserAdmin(t *testing.T) {

	type args struct {
		u     *models.User
		admin bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				us := &UserServiceImpl{
					userMap:    tt.fields.userMap,
					userLock:   tt.fields.userLock,
					publicUser: tt.fields.publicUser,
					rootUser:   tt.fields.rootUser,
					col:        tt.fields.col,
				}
				tt.wantErr(
					t, us.SetUserAdmin(tt.args.u, tt.args.admin),
					fmt.Sprintf("SetUserAdmin(%v, %v)", tt.args.u, tt.args.admin),
				)
			},
		)
	}
}

func TestUserServiceImpl_Size(t *testing.T) {

	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				us := &UserServiceImpl{
					userMap:    tt.fields.userMap,
					userLock:   tt.fields.userLock,
					publicUser: tt.fields.publicUser,
					rootUser:   tt.fields.rootUser,
					col:        tt.fields.col,
				}
				assert.Equalf(t, tt.want, us.Size(), "Size()")
			},
		)
	}
}

func TestUserServiceImpl_UpdateUserPassword(t *testing.T) {

	type args struct {
		username      models.Username
		oldPassword   string
		newPassword   string
		allowEmptyOld bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				us := &UserServiceImpl{
					userMap:    tt.fields.userMap,
					userLock:   tt.fields.userLock,
					publicUser: tt.fields.publicUser,
					rootUser:   tt.fields.rootUser,
					col:        tt.fields.col,
				}
				tt.wantErr(
					t, us.UpdateUserPassword(
						tt.args.username, tt.args.oldPassword, tt.args.newPassword, tt.args.allowEmptyOld,
					), fmt.Sprintf(
						"UpdateUserPassword(%v, %v, %v, %v)", tt.args.username, tt.args.oldPassword,
						tt.args.newPassword, tt.args.allowEmptyOld,
					),
				)
			},
		)
	}
}
