package comm

import (
	"errors"
	"fmt"
	"iter"
	"runtime"
	"runtime/debug"
	"slices"
	"sync"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
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
	mu            sync.Mutex
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
	var weblensError error
	ok := errors.As(err, &weblensError)

	var msg WsResponseInfo
	switch ok {
	case true:
		log.ShowErr(err)
		// wsc.err(err)
		msg = WsResponseInfo{EventTag: "error", Error: err.Error()}
	case false:
		wsc.errTrace(err)
		msg = WsResponseInfo{EventTag: "error", Error: "Masked unexpected server error"}
	}

	wsc.send(msg)
}

func (wsc *WsClient) PushWeblensEvent(eventTag string) {
	msg := WsResponseInfo{
		EventTag:      eventTag,
		SubscribeKey:  SubId("WEBLENS"),
		BroadcastType: ServerEvent,
	}

	wsc.send(msg)
}

func (wsc *WsClient) PushFileUpdate(updatedFile *fileTree.WeblensFile) {
	msg := WsResponseInfo{
		EventTag:      "file_updated",
		SubscribeKey:  SubId(updatedFile.ID()),
		Content:       WsC{"fileInfo": updatedFile},
		BroadcastType: FolderSubscribe,
	}

	wsc.send(msg)
}

func (wsc *WsClient) PushTaskUpdate(task *task.Task, event string, result task.TaskResult) {
	msg := WsResponseInfo{
		EventTag:      string(event),
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
		EventTag:      string(event),
		SubscribeKey:  SubId(pool.ID()),
		Content:       WsC(result),
		TaskType:      pool.CreatedInTask().JobName(),
		BroadcastType: TaskSubscribe,
	}

	wsc.send(msg)
}

func (wsc *WsClient) GetSubscriptions() iter.Seq[Subscription] {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()
	return slices.Values(wsc.subscriptions)
}

func (wsc *WsClient) Raw(msg any) error {
	return wsc.conn.WriteJSON(msg)
}

func (wsc *WsClient) send(msg WsResponseInfo) {
	if wsc != nil {
		wsc.mu.Lock()
		defer wsc.mu.Unlock()
		err := wsc.conn.WriteJSON(msg)
		if err != nil {
			wsc.errTrace(err)
		}
	}
}

func (wsc *WsClient) unsubscribe(key SubId) {
	wsc.mu.Lock()
	subIndex := slices.IndexFunc(wsc.subscriptions, func(s Subscription) bool { return s.Key == key })
	if subIndex == -1 {
		wsc.mu.Unlock()
		return
	}
	var subToRemove Subscription
	wsc.subscriptions, subToRemove = internal.Yoink(wsc.subscriptions, subIndex)
	wsc.mu.Unlock()

	wsc.debug("Unsubscribed from", subToRemove.Type, subToRemove.Key)
}

func (wsc *WsClient) disconnect() {
	wsc.Active = false

	wsc.mu.Lock()
	err := wsc.conn.Close()
	if err != nil {
		log.ShowErr(err)
		return
	}
	wsc.mu.Unlock()
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

func (wsc *WsClient) debug(msg ...any) {
	log.WsDebug.Printf(wsc.clientMsgFormat(msg...))
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

	GetUser() *models.User
	GetRemote() *models.Instance

	Error(error)
}
