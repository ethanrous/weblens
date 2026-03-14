package db

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethanrous/weblens/modules/config"
	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/ethanrous/weblens/modules/wlog"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/viccon/sturdyc"
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

// cachingTestContext implements context_mod.Z with real caching support for tests.
type cachingTestContext struct {
	context.Context

	db        *mongo.Database
	logger    *zerolog.Logger
	caches    map[string]*sturdyc.Client[any]
	cacheLock sync.RWMutex
}

func (c *cachingTestContext) Database() *mongo.Database {
	return c.db
}

func (c *cachingTestContext) GetCache(col string) *sturdyc.Client[any] {
	c.cacheLock.RLock()
	cache, ok := c.caches[col]
	c.cacheLock.RUnlock()

	if ok {
		return cache
	}

	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	cache = sturdyc.New[any](1000, 10, time.Hour, 10)
	c.caches[col] = cache

	return cache
}

func (c *cachingTestContext) Log() *zerolog.Logger {
	return c.logger
}

func (c *cachingTestContext) WithLogger(_ zerolog.Logger) {}

func (c *cachingTestContext) WithContext(ctx context.Context) context.Context {
	return &cachingTestContext{
		Context: ctx,
		db:      c.db,
		logger:  c.logger,
		caches:  c.caches,
	}
}

func (c *cachingTestContext) Value(key any) any {
	if key == DatabaseContextKey {
		return c.db
	}

	if key == context_mod.WgKey {
		return &sync.WaitGroup{}
	}

	return c.Context.Value(key)
}

var _ context_mod.Z = &cachingTestContext{}

// SetupTestDBWithCache creates a test database context with caching enabled.
// This returns a context that implements context_mod.Z so that the
// ContextualizedCollection caching code path can be exercised.
func SetupTestDBWithCache(t *testing.T, collectionKey string, indexModels ...mongo.IndexModel) context.Context {
	origDoCache := config.GetConfig().DoCache

	config.SetDoCache(true)

	t.Cleanup(func() {
		config.SetDoCache(origDoCache)
	})

	plainCtx := context.WithValue(context.Background(), context_mod.WgKey, &sync.WaitGroup{})

	testDB, err := ConnectToMongo(plainCtx, config.GetMongoDBUri(), safeTestName(t.Name()))
	if err != nil {
		panic(err)
	}

	logger := wlog.NewZeroLogger()

	ctx := &cachingTestContext{
		Context: context.Background(),
		db:      testDB,
		logger:  logger,
		caches:  make(map[string]*sturdyc.Client[any]),
	}

	col, err := GetCollection[any](ctx, collectionKey)
	require.NoError(t, err)

	err = col.Drop(ctx)
	require.NoError(t, err)

	for _, indexModel := range indexModels {
		_, err = col.GetCollection().Indexes().CreateOne(ctx, indexModel)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	t.Cleanup(func() {
		err := col.Drop(ctx)
		if err != nil {
			t.Logf("Failed to cleanup test collection: %v", err)
		}
	})

	return ctx
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
