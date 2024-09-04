package models_test

import (
	"testing"

	. "github.com/ethrousseau/weblens/models"
	"github.com/stretchr/testify/assert"
)

var username = "bob"
var password = "b0bz!23"

func TestUserPassword(t *testing.T) {
	t.Parallel()
	
	u, err := NewUser(Username(username), password, false, false)
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
