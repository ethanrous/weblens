package websocket

// WsAction represents the type of action in a WebSocket message sent from the client.
type WsAction string

// WsEvent represents the type of event in a WebSocket message sent to the client.
type WsEvent string

// ClientType represents the type of client connected via WebSocket.
type ClientType string

// SubscriptionType represents the type of subscription a client can have for WebSocket events.
type SubscriptionType string

// WebClient and TowerClient represent the different types of clients that can connect via WebSocket.
const (
	WebClient   ClientType = "webClient"
	TowerClient ClientType = "towerClient"
)

// All Websocket action tags. These are used to identify the type of content being sent *from* the client.
const (
	CancelTask    WsAction = "cancelTask"
	ReportError   WsAction = "showWebError"
	ScanDirectory WsAction = "scanDirectory"

	ActionSubscribe   WsAction = "subscribe"
	ActionUnsubscribe WsAction = "unsubscribe"

	// UserSubscribe does not actually get "subscribed" to, it is automatically tracked for every websocket
	// connection made, and only sends updates to that specific user when needed
	UserSubscribe SubscriptionType = "userSubscribe"

	SystemSubscribe SubscriptionType = "systemSubscribe"

	FolderSubscribe   SubscriptionType = "folderSubscribe"
	TaskSubscribe     SubscriptionType = "taskSubscribe"
	TaskTypeSubscribe SubscriptionType = "taskTypeSubscribe"
)

// SystemSubscriberKey is the key used for system-wide WebSocket subscriptions.
const SystemSubscriberKey = "WEBLENS"

// All Websocket event tags. These are used to identify the type of content being sent *to* the client.
const (
	BackupCompleteEvent          WsEvent = "backupComplete"
	BackupFailedEvent            WsEvent = "backupFailed"
	BackupProgressEvent          WsEvent = "backupProgress"
	CopyFileCompleteEvent        WsEvent = "copyFileComplete"
	CopyFileFailedEvent          WsEvent = "copyFileFailed"
	CopyFileStartedEvent         WsEvent = "copyFileStarted"
	ErrorEvent                   WsEvent = "error"
	FileCreatedEvent             WsEvent = "fileCreated"
	FileDeletedEvent             WsEvent = "fileDeleted"
	FileMovedEvent               WsEvent = "fileMoved"
	FileScanStartedEvent         WsEvent = "fileScanStarted"
	FileScanCompleteEvent        WsEvent = "fileScanComplete"
	FileScanFailedEvent          WsEvent = "fileScanFailedEvent"
	FileScanCancelledEvent       WsEvent = "fileScanCancelledEvent"
	FileUpdatedEvent             WsEvent = "fileUpdated"
	FilesDeletedEvent            WsEvent = "filesDeleted"
	FilesMovedEvent              WsEvent = "filesMoved"
	FilesUpdatedEvent            WsEvent = "filesUpdated"
	FolderScanCompleteEvent      WsEvent = "folderScanComplete"
	PoolCancelledEvent           WsEvent = "poolCancelled"
	PoolCompleteEvent            WsEvent = "poolComplete"
	PoolCreatedEvent             WsEvent = "poolCreated"
	RemoteConnectionChangedEvent WsEvent = "remoteConnectionChanged"
	RestoreCompleteEvent         WsEvent = "restoreComplete"
	RestoreFailedEvent           WsEvent = "restoreFailed"
	RestoreProgressEvent         WsEvent = "restoreProgress"
	RestoreStartedEvent          WsEvent = "restoreStarted"
	ScanDirectoryProgressEvent   WsEvent = "scanDirectoryProgress"
	ServerGoingDownEvent         WsEvent = "goingDown"
	ShareUpdatedEvent            WsEvent = "shareUpdated"
	StartupProgressEvent         WsEvent = "startupProgress"
	TaskCanceledEvent            WsEvent = "taskCanceled"
	TaskCompleteEvent            WsEvent = "taskComplete"
	TaskCreatedEvent             WsEvent = "taskCreated"
	TaskFailedEvent              WsEvent = "taskFailure"
	WeblensLoadedEvent           WsEvent = "weblensLoaded"
	ZipCompleteEvent             WsEvent = "zipComplete"
	ZipProgressEvent             WsEvent = "createZipProgress"

	// FlushEvent is used to flush the pending websocket notifications, typically after a large amount of data has been sent.
	// It is internal to the websocket module and should not be sent to clients, and clients are not required to handle it.
	FlushEvent WsEvent = "flush"
)
