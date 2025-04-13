package user

import (
	"context"
	"testing"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/stretchr/testify/assert"
)

var username = "bob"
var password = "b0bz!23"
var fullName = "Bob Smith"

func TestUserPassword(t *testing.T) {
	t.Parallel()

	u := &User{
		Username:    username,
		Password:    password,
		DisplayName: fullName,
	}

	mongodb, err := db.ConnectToMongo(context.Background(), config.GetMongoDBUri(), "weblensTestDB")
	if err != nil {
		t.Error(err)
	}
	ctx := context.WithValue(context.Background(), db.DatabaseContextKey, mongodb)

	err = CreateUser(ctx, u)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// u.Password is the hash of the password, they should not match
	assert.NotEqual(t, u.Password, password)

	wrongPassCheck := u.CheckLogin("wrongPassword")
	assert.False(t, wrongPassCheck)

	wrongPassCheck2 := u.CheckLogin(password)
	assert.False(t, wrongPassCheck2)

	u.Activated = true
	rightPassCheck := u.CheckLogin(password)
	assert.True(t, rightPassCheck)
}
