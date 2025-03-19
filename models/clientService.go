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
	UnsubTask(taskId task.Id)

	Subscribe(c *WsClient, key SubId, action WsAction, subTime time.Time, share Share) (
		complete bool, results map[task.TaskResultKey]any, err error,
	)
	Unsubscribe(c *WsClient, key SubId, unSubTime time.Time) error
}

type BasicCaster interface {
	PushWeblensEvent(eventTag string, content ...WsC)

	PushFileUpdate(updatedFile *fileTree.WeblensFileImpl, media *Media)
	PushTaskUpdate(task *task.Task, event string, result task.TaskResult)
	PushPoolUpdate(pool task.Pool, event string, result task.TaskResult)
}

type Broadcaster interface {
	BasicCaster
	PushFileCreate(newFile *fileTree.WeblensFileImpl)
	PushFileMove(preMoveFile *fileTree.WeblensFileImpl, postMoveFile *fileTree.WeblensFileImpl)
	PushFilesMove(preMoveParentId, postMoveParentId fileTree.FileId, files []*fileTree.WeblensFileImpl)
	PushFileDelete(deletedFile *fileTree.WeblensFileImpl)
	PushFilesDelete(deletedFiles []*fileTree.WeblensFileImpl)
	PushFilesUpdate(files []*fileTree.WeblensFileImpl, medias []*Media)
	PushShareUpdate(username Username, newShareInfo Share)
	Enable()
	Disable()
	IsEnabled() bool
	IsBuffered() bool

	FolderSubToTask(folder fileTree.FileId, taskId task.Id)
	// UnsubTask(task *task.Task)
	DisableAutoFlush()
	AutoFlushEnable()
	Flush()

	Relay(msg WsResponseInfo)

	// Close flush, release the auto-flusher, and disable the caster
	Close()
}

// WsC is the generic WebSocket Content container
type WsC map[string]any
type SubId = string

// WsAction is an action sent by the client to the server
type WsAction string
type ClientType string

const (
	// UserSubscribe does not actually get "subscribed" to, it is automatically tracked for every websocket
	// connection made, and only sends updates to that specific user when needed
	UserSubscribe WsAction = "userSubscribe"

	FolderSubscribe   WsAction = "folderSubscribe"
	TaskSubscribe     WsAction = "taskSubscribe"
	TaskTypeSubscribe WsAction = "taskTypeSubscribe"
	Unsubscribe       WsAction = "unsubscribe"
	ScanDirectory     WsAction = "scanDirectory"
	CancelTask        WsAction = "cancelTask"
	ReportError       WsAction = "showWebError"
)

const (
	WebClient      ClientType = "webClient"
	InstanceClient ClientType = "remoteClient"
)

type Subscription struct {
	When time.Time
	Type WsAction
	Key  SubId
}

type WsResponseInfo struct {
	EventTag      string     `json:"eventTag"`
	SubscribeKey  SubId      `json:"subscribeKey"`
	TaskType      string     `json:"taskType,omitempty"`
	Content       WsC        `json:"content"`
	Error         string     `json:"error,omitempty"`
	BroadcastType WsAction   `json:"broadcastType,omitempty"`
	RelaySource   InstanceId `json:"relaySource,omitempty"`
	SentTime      int64      `json:"sentTime,omitempty"`
}

type WsRequestInfo struct {
	Action  WsAction `json:"action"`
	Content string   `json:"content"`
	SentAt  int64    `json:"sentAt"`
}

// WsR WebSocket Request interface
type WsR interface {
	GetKey() SubId
	Action() WsAction
	GetShare(ShareService) *FileShare
}

// All Websocket event tags. These are used to identify the type of content being sent to the client
const (
	BackupCompleteEvent          = "backupComplete"
	BackupFailedEvent            = "backupFailed"
	BackupProgressEvent          = "backupProgress"
	CopyFileCompleteEvent        = "copyFileComplete"
	CopyFileFailedEvent          = "copyFileFailed"
	CopyFileStartedEvent         = "copyFileStarted"
	ErrorEvent                   = "error"
	FileCreatedEvent             = "fileCreated"
	FileDeletedEvent             = "fileDeleted"
	FileMovedEvent               = "fileMoved"
	FileScanStartedEvent         = "fileScanStarted"
	FileScanCompleteEvent        = "fileScanComplete"
	FileUpdatedEvent             = "fileUpdated"
	FilesDeletedEvent            = "filesDeleted"
	FilesMovedEvent              = "filesMoved"
	FilesUpdatedEvent            = "filesUpdated"
	FolderScanCompleteEvent      = "folderScanComplete"
	PoolCancelledEvent           = "poolCancelled"
	PoolCompleteEvent            = "poolComplete"
	PoolCreatedEvent             = "poolCreated"
	RemoteConnectionChangedEvent = "remoteConnectionChanged"
	RestoreCompleteEvent         = "restoreComplete"
	RestoreFailedEvent           = "restoreFailed"
	RestoreProgressEvent         = "restoreProgress"
	RestoreStartedEvent          = "restoreStarted"
	ScanDirectoryProgressEvent   = "scanDirectoryProgress"
	ServerGoingDownEvent         = "goingDown"
	ShareUpdatedEvent            = "shareUpdated"
	StartupProgressEvent         = "startupProgress"
	TaskCanceledEvent            = "taskCanceled"
	TaskCompleteEvent            = "taskComplete"
	TaskCreatedEvent             = "taskCreated"
	TaskFailedEvent              = "taskFailure"
	WeblensLoadedEvent           = "weblensLoaded"
	ZipCompleteEvent             = "zipComplete"
	ZipProgressEvent             = "createZipProgress"
)
