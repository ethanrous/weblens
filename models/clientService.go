package models

import (
	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/models/client"
	"github.com/ethanrous/weblens/task"
)

type ClientManager interface {
	// ClientConnect(conn *websocket.Conn, user *User) *WsClient
	// ClientDisconnect(c *WsClient)
	// RemoteConnect(conn *websocket.Conn, remote *Instance) *WsClient
	//
	// Send(msg client.WsResponseInfo)
	//
	// GetSubscribers(st client.WsAction, key string) (clients []*client.WsClient)
	// GetClientByUsername(username string) *client.WsClient
	// GetClientByServerId(id InstanceId) *WsClient
	// GetAllClients() []*WsClient
	// GetConnectedAdmins() []*WsClient
	//
	// // FolderSubToPool(folderId fileTree.FileId, poolId task.Id)
	// // TaskSubToPool(taskId task.Id, poolId task.Id)
	// FolderSubToTask(folderId fileTree.FileId, taskId task.Id)
	// UnsubTask(taskId task.Id)
	//
	// Subscribe(c *WsClient, key SubId, action WsAction, subTime time.Time, share Share) (
	//
	//	complete bool, results map[task.TaskResultKey]any, err error,
	//
	// )
	// Unsubscribe(c *WsClient, key SubId, unSubTime time.Time) error
}

type BasicCaster interface {
	// PushWeblensEvent(eventTag string, content ...WsC)
	//
	// PushFileUpdate(updatedFile *fileTree.WeblensFileImpl, media *Media)
	// PushTaskUpdate(task *task.Task, event string, result task.TaskResult)
	// PushPoolUpdate(pool task.Pool, event string, result task.TaskResult)
}

type Broadcaster interface {
	BasicCaster
	PushFileCreate(newFile *fileTree.WeblensFileImpl)
	PushFileMove(preMoveFile *fileTree.WeblensFileImpl, postMoveFile *fileTree.WeblensFileImpl)
	PushFilesMove(preMoveParentId, postMoveParentId fileTree.FileId, files []*fileTree.WeblensFileImpl)
	PushFileDelete(deletedFile *fileTree.WeblensFileImpl)
	PushFilesDelete(deletedFiles []*fileTree.WeblensFileImpl)
	PushFilesUpdate(files []*fileTree.WeblensFileImpl, medias []*Media)
	PushShareUpdate(username string, newShareInfo Share)
	Enable()
	Disable()
	IsEnabled() bool
	IsBuffered() bool

	FolderSubToTask(folder fileTree.FileId, taskId task.Id)
	// UnsubTask(task *task.Task)
	DisableAutoFlush()
	AutoFlushEnable()
	Flush()

	Relay(msg client.WsResponseInfo)

	// Close flush, release the auto-flusher, and disable the caster
	Close()
}
