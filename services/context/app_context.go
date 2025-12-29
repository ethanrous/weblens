// Package context provides application context and dependency injection for Weblens services.
package context

import (
	"context"
	"sync"
	"time"

	"github.com/ethanrous/weblens/models/client"
	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/file"
	task_model "github.com/ethanrous/weblens/models/task"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/ethanrous/weblens/modules/websocket"
	"github.com/viccon/sturdyc"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	context_mod.ToZ = func(ctx context.Context) context_mod.Z {
		if ctx == nil {
			return nil
		}

		c, ok := ctx.Value(appContextKey{}).(AppContext)
		if !ok {
			return nil
		}

		return c
	}
}

var _ context_mod.Z = AppContext{}

// ErrNoContext is returned when a context is not an AppContext.
var ErrNoContext = errors.New("context is not an AppContext")

type appContextKey struct{}

// AppContext represents the application-level context with services and resources for the Weblens application.
type AppContext struct {
	BasicContext

	// LocalTowerID is the id of the tower that the app is running on
	LocalTowerID string

	FileService   file.Service
	TaskService   *task_model.WorkerPool
	ClientService client.Manager
	DB            *mongo.Database

	Cache     map[string]*sturdyc.Client[any]
	cacheLock *sync.RWMutex

	WG *sync.WaitGroup
}

var capacity = 10000
var numShards = 10
var ttl = time.Hour
var evictionPercentage = 10

// NewAppContext creates a new AppContext from a BasicContext with initialized services and caches.
func NewAppContext(ctx BasicContext) AppContext {
	newCtx := AppContext{
		BasicContext: ctx,
		Cache:        make(map[string]*sturdyc.Client[any]),
		cacheLock:    &sync.RWMutex{},
		WG:           &sync.WaitGroup{},
	}
	newCtx.BasicContext = newCtx.BasicContext.WithValue(appContextKey{}, newCtx)

	return newCtx
}

// FromContext extracts an AppContext from a standard context.Context.
func FromContext(ctx context.Context) (AppContext, bool) {
	if ctx == nil {
		return AppContext{}, false
	}

	c, ok := ctx.Value(appContextKey{}).(AppContext)
	if !ok {
		return AppContext{}, false
	}

	c.BasicContext = NewBasicContext(ctx, c.Log())

	return c, true
}

// WithValue returns a copy of AppContext with the specified key-value pair added.
func (c AppContext) WithValue(key, value any) AppContext {
	c.BasicContext = c.BasicContext.WithValue(key, value)

	return c
}

// WithContext creates a new context by combining the AppContext with the provided context.
func (c AppContext) WithContext(ctx context.Context) context.Context {
	l, ok := log.FromContextOk(ctx)
	if !ok {
		l = log.FromContext(c)
	}

	c.BasicContext = NewBasicContext(ctx, l)

	return c
}

// Value retrieves the value associated with the given key from the AppContext.
func (c AppContext) Value(key any) any {
	if key == (appContextKey{}) {
		return c
	}

	if key == db.DatabaseContextKey {
		return c.DB
	}

	if key == context_mod.WgKey {
		return c.WG
	}

	return c.BasicContext.Value(key)
}

// WithMongoSession associates a MongoDB session context with the AppContext.
func (c AppContext) WithMongoSession(_ mongo.SessionContext) {}

// GetMongoSession retrieves the MongoDB session context associated with the AppContext.
func (c AppContext) GetMongoSession() mongo.SessionContext { return nil }

// Notify sends websocket notifications to connected clients.
func (c AppContext) Notify(ctx context.Context, data ...websocket.WsResponseInfo) {
	c.ClientService.Notify(ctx, data...)
}

// Database returns the MongoDB database instance for the application.
func (c AppContext) Database() *mongo.Database {
	return c.DB
}

// GetCache retrieves or creates a cache client for the specified collection.
func (c AppContext) GetCache(col string) *sturdyc.Client[any] {
	c.cacheLock.RLock()

	cache, ok := c.Cache[col]
	if ok {
		c.cacheLock.RUnlock()

		return cache
	}

	c.cacheLock.RUnlock()

	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	cache = sturdyc.New[any](capacity, numShards, ttl, evictionPercentage)
	c.Cache[col] = cache

	return cache
}

// ClearCache removes all cached data from the AppContext.
func (c AppContext) ClearCache() {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	for k := range c.Cache {
		delete(c.Cache, k)
	}
}

// DispatchJob submits a new job to the task service for asynchronous execution.
func (c AppContext) DispatchJob(jobName string, meta task_mod.Metadata, pool task_mod.Pool) (task_mod.Task, error) {
	return c.TaskService.DispatchJob(c, jobName, meta, pool)
}

// GetFileService returns the file service instance from the AppContext.
func (c AppContext) GetFileService() file.Service {
	return c.FileService
}

// GetTowerID returns the ID of the local tower.
func (c AppContext) GetTowerID() string {
	return c.LocalTowerID
}
