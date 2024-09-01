package database_test

import (
	"testing"

	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/internal"
	"github.com/stretchr/testify/assert"
)

func TestConnectToMongo(t *testing.T) {
	mondb, err := database.ConnectToMongo(internal.GetMongoURI(), internal.GetMongoDBName())
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.NotNil(t, mondb)

	mondb, err = database.ConnectToMongo("notmongo:22000", "notaamongodb")
	assert.Error(t, err)
	assert.Nil(t, mondb)
}
