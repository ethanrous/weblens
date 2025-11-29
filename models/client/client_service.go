package client

import (
	"context"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	task_model "github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/gorilla/websocket"
)

type ClientManager interface {
	Notify(ctx context.Context, msg ...websocket_mod.WsResponseInfo)
	ClientConnect(ctx context_mod.LoggerContext, conn *websocket.Conn, user *user_model.User) (*WsClient, error)
	RemoteConnect(ctx context_mod.LoggerContext, conn *websocket.Conn, remote *tower_model.Instance) *WsClient
	ClientDisconnect(ctx context.Context, c *WsClient)
	DisconnectAll(ctx context.Context)
	GetClientByUsername(username string) *WsClient
	GetClientByTowerId(towerId string) *WsClient
	GetAllClients() []*WsClient
	GetConnectedAdmins() []*WsClient
	GetSubscribers(ctx context_mod.LoggerContext, st websocket_mod.SubscriptionType, key string) (clients []*WsClient)
	SubscribeToFile(ctx context_mod.ContextZ, c *WsClient, file *file_model.WeblensFileImpl, share *share_model.FileShare, subTime time.Time) error
	SubscribeToTask(ctx context_mod.LoggerContext, c *WsClient, task *task_model.Task, subTime time.Time) error
	Unsubscribe(ctx context_mod.LoggerContext, c *WsClient, key string, unSubTime time.Time) error
	FolderSubToTask(ctx context_mod.LoggerContext, folderId string, task task.Task)
	UnsubTask(ctx context.Context, taskId string)
	Send(ctx context.Context, msg websocket_mod.WsResponseInfo)
	Flush(ctx context.Context)
	Relay(msg websocket_mod.WsResponseInfo)
	// PushWeblensEvent(event websocket_mod.WsEvent, msg websocket_mod.WsData)
}
