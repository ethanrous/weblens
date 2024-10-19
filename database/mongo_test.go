package database_test

import (
	"testing"

	"github.com/ethanrous/weblens/database"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectToMongo(t *testing.T) {
	mondb, err := database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName())
	require.NoError(t, err)

	assert.NotNil(t, mondb)

	mondb, err = database.ConnectToMongo("notmongo:22000", "notaamongodb")
	assert.Error(t, err)
	assert.Nil(t, mondb)
}
