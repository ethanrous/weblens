package client

import (
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	task_model "github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/gorilla/websocket"
)

type ClientManager interface {
	Notify(msg ...websocket_mod.WsResponseInfo)
	ClientConnect(ctx context.LoggerContext, conn *websocket.Conn, user *user_model.User) (*WsClient, error)
	RemoteConnect(ctx context.LoggerContext, conn *websocket.Conn, remote *tower_model.Instance) *WsClient
	ClientDisconnect(ctx context.LoggerContext, c *WsClient)
	GetClientByUsername(username string) *WsClient
	GetClientByServerId(instanceId string) *WsClient
	GetAllClients() []*WsClient
	GetConnectedAdmins() []*WsClient
	GetSubscribers(ctx context.LoggerContext, st websocket_mod.WsAction, key string) (clients []*WsClient)
	SubscribeToFile(ctx context.ContextZ, c *WsClient, file *file_model.WeblensFileImpl, share *share_model.FileShare, subTime time.Time) error
	SubscribeToTask(ctx context.LoggerContext, c *WsClient, task *task_model.Task, subTime time.Time) error
	Unsubscribe(ctx context.LoggerContext, c *WsClient, key string, unSubTime time.Time) error
	FolderSubToTask(ctx context.LoggerContext, folderId string, task task.Task)
	UnsubTask(ctx context.LoggerContext, taskId string)
	Send(ctx context.LoggerContext, msg websocket_mod.WsResponseInfo)
	Relay(msg websocket_mod.WsResponseInfo)
	// PushWeblensEvent(event websocket_mod.WsEvent, msg websocket_mod.WsData)
}
