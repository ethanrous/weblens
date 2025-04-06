package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	context_service "github.com/ethanrous/weblens/services/context"

	"github.com/ethanrous/weblens/models/db"
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

	ctx := NewDBContext(t.Name())

	err := CreateUser(ctx, u)
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

func NewDBContext(dbName string) context_service.AppContext {
	ctx := context_service.BasicContext{
		Context: context.Background(),
	}

	db.ConnectToMongo(ctx, "", dbName)

	return context_service.AppContext{
		BasicContext: ctx, DB: nil,
	}
}
