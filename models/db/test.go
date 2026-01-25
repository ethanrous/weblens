package db

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/ethanrous/weblens/modules/config"
	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func safeTestName(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	strings.ReplaceAll(name, ".", "-")
	strings.ReplaceAll(name, "_", "")

	// limit name length to 63 characters
	name = name[:min(len(name), 63)]

	return name
}

// SetupTestDB creates a test database context with a clean collection and optional indexes.
func SetupTestDB(t *testing.T, collectionKey string, indexModels ...mongo.IndexModel) context.Context {
	ctx := context.WithValue(context.Background(), context_mod.WgKey, &sync.WaitGroup{})

	testDB, err := ConnectToMongo(ctx, config.GetMongoDBUri(), safeTestName(t.Name()))
	if err != nil {
		panic(err)
	}

	// Set the MongoDB instance in the context
	// Connecting to mongo with test context would cancel before the cleanup is run. We must use our own context and cancel.
	ctx = context.WithValue(context.Background(), DatabaseContextKey, testDB) //nolint:revive

	// Clean up test collection before each test
	col, err := GetCollection[any](ctx, collectionKey)
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
	}

	t.Cleanup(cleanup)

	return ctx
}
