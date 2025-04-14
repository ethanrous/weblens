package context

import (
	"github.com/ethanrous/weblens/models/client"
	"github.com/ethanrous/weblens/models/file"
	task_model "github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/context"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/ethanrous/weblens/modules/websocket"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ context.ContextZ = &AppContext{}

type AppContext struct {
	BasicContext

	// LocalTowerId is the id of the tower that the app is running on
	LocalTowerId string

	FileService   file.FileService
	TaskService   *task_model.WorkerPool
	ClientService client.ClientManager
	DB            *mongo.Database
}

func NewAppContext(ctx *BasicContext) *AppContext {
	return &AppContext{BasicContext: *ctx}
}

func (c *AppContext) AppCtx() context.ContextZ {
	return c
}

func (c *AppContext) WithLogger(l *zerolog.Logger) *AppContext {
	newL := l.With().Logger()
	c.Logger = &newL
	return c
}

func (c *AppContext) WithMongoSession(session mongo.SessionContext) {}

func (c *AppContext) GetMongoSession() mongo.SessionContext { return nil }

func (c *AppContext) Notify(data ...websocket.WsResponseInfo) {
	c.ClientService.Notify(data...)
}

func (c *AppContext) Database() *mongo.Database {
	return c.DB
}

func (c *AppContext) DispatchJob(jobName string, meta task_mod.TaskMetadata, pool task_mod.Pool) (task_mod.Task, error) {
	c.TaskService.DispatchJob(c, jobName, meta, pool.(*task_model.TaskPool))
	return nil, nil
}

func (c *AppContext) GetFileService() file.FileService {
	return c.FileService
}
