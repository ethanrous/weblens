package comm

import (
	"fmt"
	"iter"
	"runtime"
	"runtime/debug"
	"slices"
	"sync"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
	"github.com/gorilla/websocket"
)

type ClientId string

var _ Client = (*WsClient)(nil)

type WsClient struct {
	Active        bool
	connId        ClientId
	conn          *websocket.Conn
	updateMu sync.Mutex
	subsMu   sync.Mutex
	subscriptions []Subscription
	user          *models.User
	remote        *models.Instance
}

func (wsc *WsClient) IsOpen() bool {
	return wsc.Active
}

func (wsc *WsClient) GetClientId() ClientId {
	return wsc.connId
}

func (wsc *WsClient) ClientType() ClientType {
	return WebClient
}

func (wsc *WsClient) GetShortId() ClientId {
	if wsc.connId == "" {
		return ""
	}
	return wsc.connId[28:]
}

func (wsc *WsClient) GetUser() *models.User {
	return wsc.user
}

func (wsc *WsClient) GetRemote() *models.Instance {
	return wsc.remote
}

func (wsc *WsClient) ReadOne() (int, []byte, error) {
	return wsc.conn.ReadMessage()
}

func (wsc *WsClient) Error(err error) {
	safe, _ := werror.TrySafeErr(err)
	wsc.send(WsResponseInfo{EventTag: "error", Error: safe.Error()})
}

func (wsc *WsClient) PushWeblensEvent(eventTag string) {
	msg := WsResponseInfo{
		EventTag:      eventTag,
		SubscribeKey:  SubId("WEBLENS"),
		BroadcastType: ServerEvent,
	}

	wsc.send(msg)
}

func (wsc *WsClient) PushFileUpdate(updatedFile *fileTree.WeblensFile, media *models.Media) {
	msg := WsResponseInfo{
		EventTag:      "file_updated",
		SubscribeKey:  SubId(updatedFile.ID()),
		Content: WsC{"fileInfo": updatedFile, "mediaData": media},
		BroadcastType: FolderSubscribe,
	}

	wsc.send(msg)
}

func (wsc *WsClient) PushTaskUpdate(task *task.Task, event string, result task.TaskResult) {
	msg := WsResponseInfo{
		EventTag: event,
		SubscribeKey:  SubId(task.TaskId()),
		Content:       WsC(result),
		TaskType:      task.JobName(),
		BroadcastType: TaskSubscribe,
	}

	wsc.send(msg)
}

func (wsc *WsClient) PushPoolUpdate(pool task.Pool, event string, result task.TaskResult) {
	if pool.IsGlobal() {
		log.Warning.Println("Not pushing update on global pool")
		return
	}

	msg := WsResponseInfo{
		EventTag: event,
		SubscribeKey:  SubId(pool.ID()),
		Content:       WsC(result),
		TaskType:      pool.CreatedInTask().JobName(),
		BroadcastType: TaskSubscribe,
	}

	wsc.send(msg)
}

func (wsc *WsClient) GetSubscriptions() iter.Seq[Subscription] {
	wsc.updateMu.Lock()
	defer wsc.updateMu.Unlock()
	return slices.Values(wsc.subscriptions)
}

func (wsc *WsClient) Raw(msg any) error {
	return wsc.conn.WriteJSON(msg)
}

func (wsc *WsClient) SubLock() {
	wsc.subsMu.Lock()
}

func (wsc *WsClient) SubUnlock() {
	wsc.subsMu.Unlock()
}

func (wsc *WsClient) send(msg WsResponseInfo) {
	if wsc != nil && wsc.Active {
		wsc.updateMu.Lock()
		defer wsc.updateMu.Unlock()
		err := wsc.conn.WriteJSON(msg)
		if err != nil {
			wsc.errTrace(err)
		}
	} else {
		log.Error.Println("Trying to send to closed client")
	}
}

func (wsc *WsClient) unsubscribe(key SubId) {
	wsc.updateMu.Lock()
	subIndex := slices.IndexFunc(wsc.subscriptions, func(s Subscription) bool { return s.Key == key })
	if subIndex == -1 {
		wsc.updateMu.Unlock()
		return
	}
	var subToRemove Subscription
	wsc.subscriptions, subToRemove = internal.Yoink(wsc.subscriptions, subIndex)
	wsc.updateMu.Unlock()

	log.Debug.Printf("[%s] unsubscribing from %s", wsc.user.GetUsername(), subToRemove)
}

func (wsc *WsClient) disconnect() {
	wsc.Active = false

	wsc.updateMu.Lock()
	err := wsc.conn.Close()
	if err != nil {
		log.ShowErr(err)
		return
	}
	wsc.updateMu.Unlock()
	wsc.log("Disconnected")
}

func (wsc *WsClient) clientMsgFormat(msg ...any) string {
	var clientName string
	if wsc.GetUser() != nil {
		clientName = string(wsc.GetUser().GetUsername())
	} else {
		clientName = wsc.GetRemote().GetName()
	}
	return fmt.Sprintf("| %s (%s) | %s", wsc.GetShortId(), clientName, fmt.Sprintln(msg...))
}

func (wsc *WsClient) log(msg ...any) {
	log.WsInfo.Printf(wsc.clientMsgFormat(msg...))
}

func (wsc *WsClient) err(msg ...any) {
	_, file, line, _ := runtime.Caller(2)
	msg = []any{any(fmt.Sprintf("%s:%d:", file, line)), msg}
	log.WsError.Printf(wsc.clientMsgFormat(msg...))
}

func (wsc *WsClient) errTrace(msg ...any) {
	_, file, line, _ := runtime.Caller(2)
	msg = []any{any(fmt.Sprintf("%s:%d:", file, line)), msg}
	log.WsError.Printf(wsc.clientMsgFormat(msg...), string(debug.Stack()))
}

type ClientManager interface {
	ClientConnect(conn *websocket.Conn, user *models.User) *WsClient
	RemoteConnect(conn *websocket.Conn, remote *models.Instance) *WsClient
	GetSubscribers(st WsAction, key SubId) (clients []*WsClient)
	GetClientByUsername(username models.Username) *WsClient
	GetClientByInstanceId(id models.InstanceId) *WsClient
	GetAllClients() []*WsClient
	GetConnectedAdmins() []*WsClient

	FolderSubToPool(folderId fileTree.FileId, poolId task.TaskId)

	Subscribe(c *WsClient, key SubId, action WsAction, share models.Share) (
		complete bool,
		results map[string]any, err error,
	)
	Unsubscribe(c *WsClient, key SubId) error

	Send(msg WsResponseInfo)

	ClientDisconnect(c *WsClient)
}

type Client interface {
	BasicCaster

	IsOpen() bool

	ReadOne() (int, []byte, error)

	GetSubscriptions() iter.Seq[Subscription]
	GetClientId() ClientId
	GetShortId() ClientId

	SubLock()
	SubUnlock()

	GetUser() *models.User
	GetRemote() *models.Instance

	Error(error)
}
