package db

import (
	"context"
	"testing"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

var testDB *mongo.Database

func initMongo() {
	if testing.Testing() {
		var err error

		testDB, err = ConnectToMongo(context.Background(), config.GetMongoDBUri(), "weblensTestDB")
		if err != nil {
			panic(err)
		}
	} else {
		panic("MongoDB should not be initialized in non-testing mode")
	}
}

func SetupTestDB(t *testing.T, collectionKey string, indexModels ...mongo.IndexModel) context.Context {
	// ctx, cancel := context.WithCancel(context.Background())
	// mongodb, err := ConnectToMongo(ctx, config.GetMongoDBUri(), "weblensTestDB")
	// require.NoError(t, err)

	if testDB == nil {
		initMongo()
	}

	// Set the MongoDB instance in the context
	// Connecting to mongo with test context would cancel before the cleanup is run. We must use our own context and cancel.
	ctx := context.WithValue(context.Background(), DatabaseContextKey, testDB)

	// Clean up test collection before each test
	col, err := GetCollection(ctx, collectionKey)
	require.NoError(t, err)

	err = col.Drop(ctx)
	require.NoError(t, err)

	// Set indexes for the collection, if any
	for _, indexModel := range indexModels {
		_, err = col.GetCollection().Indexes().CreateOne(ctx, indexModel)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	// Install cleanup for after the test, too
	cleanup := func() {
		err := col.Drop(ctx)
		if err != nil {
			t.Logf("Failed to cleanup test collection: %v", err)
		}

		// Close the MongoDB connection
		// cancel()
	}

	t.Cleanup(cleanup)

	return ctx
}
