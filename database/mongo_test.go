package database_test

import (
	"testing"

	"github.com/ethanrous/weblens/database"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectToMongo(t *testing.T) {
	logger := log.NewZeroLogger()
	mondb, err := database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName(env.Config{}), logger)
	require.NoError(t, err)

	assert.NotNil(t, mondb)

	mondb, err = database.ConnectToMongo("notmongo:22000", "notaamongodb", logger)
	assert.Error(t, err)
	assert.Nil(t, mondb)
}
