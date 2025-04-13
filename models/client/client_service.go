package client

import (
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	task_model "github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/context"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/gorilla/websocket"
)

type ClientManager interface {
	ClientConnect(ctx context.ContextZ, conn *websocket.Conn, user *user_model.User) *WsClient
	RemoteConnect(ctx context.ContextZ, conn *websocket.Conn, remote *tower_model.Instance) *WsClient
	ClientDisconnect(ctx context.ContextZ, c *WsClient)
	GetClientByUsername(username string) *WsClient
	GetClientByServerId(instanceId string) *WsClient
	GetAllClients() []*WsClient
	GetConnectedAdmins() []*WsClient
	GetSubscribers(ctx context.ContextZ, st websocket_mod.WsAction, key string) (clients []*WsClient)
	SubscribeToFile(ctx context.ContextZ, c *WsClient, file *file_model.WeblensFileImpl, share *share_model.FileShare, subTime time.Time) error
	SubscribeToTask(ctx context.ContextZ, c *WsClient, task *task_model.Task, subTime time.Time) error
	Unsubscribe(ctx context.ContextZ, c *WsClient, key string, unSubTime time.Time) error
	FolderSubToTask(ctx context.ContextZ, folderId string, taskId string)
	UnsubTask(ctx context.ContextZ, taskId string)
	Send(ctx context.ContextZ, msg websocket_mod.WsResponseInfo)
	Relay(msg websocket_mod.WsResponseInfo)
	// PushWeblensEvent(event websocket_mod.WsEvent, msg websocket_mod.WsData)
}
