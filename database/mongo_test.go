package database_test

import (
	"context"
	"sync"
	"testing"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/config"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectToMongo(t *testing.T) {
	defer tests.Recover(t)
	ctx := tests.Setup(t)

	cnf := config.GetConfig()

	ctx = context.WithValue(ctx, context_mod.WgKey, &sync.WaitGroup{})
	mondb, err := db.ConnectToMongo(ctx, cnf.MongoDBUri, cnf.MongoDBName)
	require.NoError(t, err)

	assert.NotNil(t, mondb)

	mondb, err = db.ConnectToMongo(ctx, "notmongo:22000", "notaamongodb")
	assert.Error(t, err)
	assert.Nil(t, mondb)
}
