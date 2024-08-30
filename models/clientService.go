package models

import (
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/task"
	"github.com/gorilla/websocket"
)

type ClientManager interface {
	ClientConnect(conn *websocket.Conn, user *User) *WsClient
	RemoteConnect(conn *websocket.Conn, remote *Instance) *WsClient
	GetSubscribers(st WsAction, key SubId) (clients []*WsClient)
	GetClientByUsername(username Username) *WsClient
	GetClientByInstanceId(id InstanceId) *WsClient
	GetAllClients() []*WsClient
	GetConnectedAdmins() []*WsClient

	FolderSubToPool(folderId fileTree.FileId, poolId task.TaskId)
	TaskSubToPool(taskId task.TaskId, poolId task.TaskId)

	Subscribe(c *WsClient, key SubId, action WsAction, subTime time.Time, share Share) (
		complete bool, results map[string]any, err error,
	)
	Unsubscribe(c *WsClient, key SubId, unSubTime time.Time) error

	Send(msg WsResponseInfo)

	ClientDisconnect(c *WsClient)
}
