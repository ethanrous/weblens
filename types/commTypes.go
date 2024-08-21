package types

import (
	websocket2 "github.com/ethrousseau/weblens/api/websocket"
	"github.com/gorilla/websocket"
)

type ClientManager interface {
	ClientConnect(conn *websocket.Conn, user User) Client
	RemoteConnect(conn *websocket.Conn, remote Instance) Client
	AddSubscription(subInfo Subscription, client Client)
	GetSubscribers(st WsAction, key SubId) (clients []Client)
	GetClientByUsername(Username) Client
	GetClientByInstanceId(InstanceId) Client
	GetAllClients() []Client
	GetConnectedAdmins() []Client
	RemoveSubscription(subscription Subscription, client Client, removeAll bool)

	ClientDisconnect(c Client)
}

type Client interface {
	websocket2.BasicCaster

	IsOpen() bool

	ReadOne() (int, []byte, error)

	Subscribe(key SubId, action WsAction, acc AccessMeta) (complete bool, results map[string]any)
	Unsubscribe(SubId)

	GetSubscriptions() []Subscription
	GetClientId() ClientId
	GetShortId() ClientId

	GetUser() User
	GetRemote() Instance

	Error(error)

	Disconnect()
}

// WsC is the generic WebSocket Content container
type WsC map[string]any

type ClientId string
type SubId string

type Subscription struct {
	Type WsAction
	Key  SubId
}

// type Requester interface {
// 	// AttachToCore RequestCoreSnapshot() ([]FileJournalEntry, error)
// 	// AttachToCore(Instance) (Instance, error)
// 	GetCoreUsers() (us []User, err error)
// 	PingCore() bool
// 	GetCoreFileBin(f WeblensFile) ([][]byte, error)
// 	// GetCoreFileInfos(fIds []FileId) ([]WeblensFile, error)
// }

type WsAction string

const (
	UserSubscribe WsAction = "user_subscribe" // This one does not actually get "subscribed" to, it is automatically tracked for every websocket

	FolderSubscribe WsAction = "folder_subscribe"
	ServerEvent     WsAction = "server_event"
	TaskSubscribe   WsAction = "task_subscribe"
	PoolSubscribe   WsAction = "pool_subscribe"
	TaskTypeSubscribe WsAction = "task_type_subscribe"
	Unsubscribe     WsAction = "unsubscribe"
	ScanDirectory   WsAction = "scan_directory"
	CancelTask      WsAction = "cancel_task"
)

type ClientType string

const (
	WebClient    ClientType = "webClient"
	RemoteClient ClientType = "remoteClient"
)

// WsR WebSocket Request interface
type WsR interface {
	GetKey() SubId
	Action() WsAction
}

type WsResponseInfo struct {
	EventTag      string   `json:"eventTag"`
	SubscribeKey  SubId    `json:"subscribeKey"`
	TaskType      TaskType `json:"taskType,omitempty"`
	Content       WsC      `json:"content"`
	Error         string   `json:"error,omitempty"`
	BroadcastType WsAction `json:"broadcastType,omitempty"`
}

type WsRequestInfo struct {
	Action  WsAction `json:"action"`
	Content string   `json:"content"`
}
