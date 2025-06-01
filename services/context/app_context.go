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
	context_mod.ToZ = func(ctx context.Context) context_mod.ContextZ {
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

var _ context_mod.ContextZ = AppContext{}

var ErrNoContext = errors.New("context is not an AppContext")

type appContextKey struct{}
type AppContext struct {
	BasicContext

	// LocalTowerId is the id of the tower that the app is running on
	LocalTowerId string

	FileService   file.FileService
	TaskService   *task_model.WorkerPool
	ClientService client.ClientManager
	DB            *mongo.Database
	Cache         map[string]*sturdyc.Client[any]

	cacheLock *sync.RWMutex
}

var capacity = 10000
var numShards = 10
var ttl = time.Hour
var evictionPercentage = 10

func NewAppContext(ctx BasicContext) AppContext {
	newCtx := AppContext{BasicContext: ctx, Cache: make(map[string]*sturdyc.Client[any]), cacheLock: &sync.RWMutex{}}
	newCtx.BasicContext = newCtx.BasicContext.WithValue(appContextKey{}, newCtx)

	return newCtx
}

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

func (c AppContext) WithValue(key, value any) AppContext {
	c.BasicContext = c.BasicContext.WithValue(key, value)

	return c
}

func (c AppContext) WithContext(ctx context.Context) context.Context {
	l, ok := log.FromContextOk(ctx)
	if !ok {
		l = log.FromContext(c)
	}

	c.BasicContext = NewBasicContext(ctx, l)

	return c
}

func (c AppContext) Value(key any) any {
	if key == (appContextKey{}) {
		return c
	}

	if key == db.DatabaseContextKey {
		return c.DB
	}

	return c.BasicContext.Value(key)
}

func (c AppContext) WithMongoSession(session mongo.SessionContext) {}

func (c AppContext) GetMongoSession() mongo.SessionContext { return nil }

func (c AppContext) Notify(ctx context.Context, data ...websocket.WsResponseInfo) {
	c.ClientService.Notify(ctx, data...)
}

func (c AppContext) Database() *mongo.Database {
	return c.DB
}

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

func (c AppContext) ClearCache() {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	for k := range c.Cache {
		delete(c.Cache, k)
	}
}

func (c AppContext) DispatchJob(jobName string, meta task_mod.TaskMetadata, pool task_mod.Pool) (task_mod.Task, error) {
	return c.TaskService.DispatchJob(c, jobName, meta, pool)
}

func (c AppContext) GetFileService() file.FileService {
	return c.FileService
}

func (c AppContext) GetTowerId() string {
	return c.LocalTowerId
}
