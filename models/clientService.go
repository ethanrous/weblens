package models

import (
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/task"
	"github.com/gorilla/websocket"
)

type ClientManager interface {
	ClientConnect(conn *websocket.Conn, user *User) *WsClient
	ClientDisconnect(c *WsClient)
	RemoteConnect(conn *websocket.Conn, remote *Instance) *WsClient

	Send(msg WsResponseInfo)

	GetSubscribers(st WsAction, key SubId) (clients []*WsClient)
	GetClientByUsername(username Username) *WsClient
	GetClientByServerId(id InstanceId) *WsClient
	GetAllClients() []*WsClient
	GetConnectedAdmins() []*WsClient

	// FolderSubToPool(folderId fileTree.FileId, poolId task.Id)
	// TaskSubToPool(taskId task.Id, poolId task.Id)
	FolderSubToTask(folderId fileTree.FileId, taskId task.Id)

	Subscribe(c *WsClient, key SubId, action WsAction, subTime time.Time, share Share) (
		complete bool, results map[string]any, err error,
	)
	Unsubscribe(c *WsClient, key SubId, unSubTime time.Time) error
}
