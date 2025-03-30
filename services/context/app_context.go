package context

import (
	"github.com/ethanrous/weblens/models/client"
	"github.com/ethanrous/weblens/models/file"
	task_model "github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/context"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/ethanrous/weblens/modules/websocket"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ context.ContextZ = &AppContext{}

type AppContext struct {
	BasicContext

	// TowerId is the id of the tower that the app is running on
	TowerId string

	FileService   file.FileService
	TaskService   task_model.TaskService
	ClientService client.ClientManager
	DB            *mongo.Database
}

func (c *AppContext) Notify(data ...websocket.WsResponseInfo) {

}

func (c *AppContext) Database() *mongo.Database {
	return c.DB
}

func (c *AppContext) DispatchJob(string, any) (task_mod.Task, error) {
	return nil, nil
}
