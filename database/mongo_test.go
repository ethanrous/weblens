package database_test

import (
	"testing"

	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/internal/env"
	"github.com/stretchr/testify/assert"
)

func TestConnectToMongo(t *testing.T) {
	mondb, err := database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName())
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.NotNil(t, mondb)

	mondb, err = database.ConnectToMongo("notmongo:22000", "notaamongodb")
	assert.Error(t, err)
	assert.Nil(t, mondb)
}
