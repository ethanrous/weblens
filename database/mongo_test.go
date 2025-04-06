package database_test

import (
	"testing"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestConnectToMongo(t *testing.T) {
	ctx := context.TODO()

	cnf := config.GetConfig()

	mondb, err := db.ConnectToMongo(ctx, cnf.MongoDBUri, cnf.MongoDBName)
	require.NoError(t, err)

	assert.NotNil(t, mondb)

	mondb, err = db.ConnectToMongo(ctx, "notmongo:22000", "notaamongodb")
	assert.Error(t, err)
	assert.Nil(t, mondb)
}
