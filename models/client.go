package models

import (
	"iter"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/task"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ClientId = string

var _ Client = (*WsClient)(nil)

type WsClient struct {
	Active        atomic.Bool
	connId        ClientId
	conn          *websocket.Conn
	updateMu      sync.Mutex
	subsMu        sync.Mutex
	subscriptions []Subscription
	user          *User
	remote        *Instance
}

func NewClient(conn *websocket.Conn, socketUser SocketUser) *WsClient {
	newClient := &WsClient{
		connId:   ClientId(uuid.New().String()),
		conn:     conn,
		updateMu: sync.Mutex{},
		subsMu:   sync.Mutex{},
	}
	newClient.Active.Store(true)

	if socketUser.SocketType() == "webClient" {
		newClient.user = socketUser.(*User)
	} else if socketUser.SocketType() == "serverClient" {
		newClient.remote = socketUser.(*Instance)
	}

	return newClient
}

func (wsc *WsClient) IsOpen() bool {
	return wsc.Active.Load()
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

func (wsc *WsClient) GetUser() *User {
	return wsc.user
}

func (wsc *WsClient) GetRemote() *Instance {
	return wsc.remote
}

func (wsc *WsClient) ReadOne() (int, []byte, error) {
	return wsc.conn.ReadMessage()
}

func (wsc *WsClient) Error(err error) {
	safe, _ := werror.TrySafeErr(err)
	err = wsc.Send(WsResponseInfo{EventTag: "error", Error: safe.Error()})
	log.ErrTrace(err)
}

func (wsc *WsClient) PushWeblensEvent(eventTag string, content ...WsC) {
	msg := WsResponseInfo{
		EventTag:      eventTag,
		SubscribeKey:  "WEBLENS",
		BroadcastType: "serverEvent",
		SentTime:      time.Now().UnixMilli(),
	}

	if len(content) != 0 {
		msg.Content = content[0]
	}

	log.ErrTrace(wsc.Send(msg))
}

func (wsc *WsClient) PushFileUpdate(updatedFile *fileTree.WeblensFileImpl, media *Media) {
	msg := WsResponseInfo{
		EventTag:      FileUpdatedEvent,
		SubscribeKey:  updatedFile.ID(),
		Content:       WsC{"fileInfo": updatedFile, "mediaData": media},
		BroadcastType: FolderSubscribe,
		SentTime:      time.Now().UnixMilli(),
	}

	log.ErrTrace(wsc.Send(msg))
}

func (wsc *WsClient) PushTaskUpdate(task *task.Task, event string, result task.TaskResult) {
	msg := WsResponseInfo{
		EventTag:      event,
		SubscribeKey:  task.TaskId(),
		Content:       result.ToMap(),
		TaskType:      task.JobName(),
		BroadcastType: TaskSubscribe,
		SentTime:      time.Now().UnixMilli(),
	}

	log.ErrTrace(wsc.Send(msg))
}

func (wsc *WsClient) PushPoolUpdate(pool task.Pool, event string, result task.TaskResult) {
	if pool.IsGlobal() {
		log.Warning.Println("Not pushing update on global pool")
		return
	}

	msg := WsResponseInfo{
		EventTag:      event,
		SubscribeKey:  pool.ID(),
		Content:       result.ToMap(),
		TaskType:      pool.CreatedInTask().JobName(),
		BroadcastType: TaskSubscribe,
	}

	log.ErrTrace(wsc.Send(msg))
}

func (wsc *WsClient) GetSubscriptions() iter.Seq[Subscription] {
	wsc.updateMu.Lock()
	defer wsc.updateMu.Unlock()
	return slices.Values(wsc.subscriptions)
}

func (wsc *WsClient) AddSubscription(sub Subscription) {
	wsc.updateMu.Lock()
	defer wsc.updateMu.Unlock()
	wsc.subscriptions = append(wsc.subscriptions, sub)
}

func (wsc *WsClient) RemoveSubscription(key SubId) {
	wsc.updateMu.Lock()
	subIndex := slices.IndexFunc(wsc.subscriptions, func(s Subscription) bool { return s.Key == key })
	if subIndex == -1 {
		wsc.updateMu.Unlock()
		return
	}
	var subToRemove Subscription
	wsc.subscriptions, subToRemove = internal.Yoink(wsc.subscriptions, subIndex)
	wsc.updateMu.Unlock()

	log.Trace.Func(func(l log.Logger) { l.Printf("[%s] unsubscribing from %s", wsc.user.GetUsername(), subToRemove) })
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

func (wsc *WsClient) Send(msg WsResponseInfo) error {
	if wsc != nil && wsc.Active.Load() {
		wsc.updateMu.Lock()
		defer wsc.updateMu.Unlock()

		log.Debug.Func(func(l log.Logger) { l.Printf("Sending [%s] event to client [%s]", msg.EventTag, wsc.getClientName()) })

		err := wsc.conn.WriteJSON(msg)
		if err != nil {
			return werror.WithStack(err)
		}
	} else {
		return werror.Errorf("trying to send to closed client")
	}

	return nil
}

func (wsc *WsClient) Disconnect() {
	wsc.Active.Store(false)

	wsc.updateMu.Lock()
	err := wsc.conn.Close()
	if err != nil {
		log.ShowErr(err)
		return
	}
	wsc.updateMu.Unlock()

	log.Trace.Func(func(l log.Logger) { l.Printf("Disconnected %s client [%s]", wsc.getClientType(), wsc.getClientName()) })
}

func (wsc *WsClient) getClientName() string {
	if wsc.remote != nil {
		return wsc.remote.GetName()
	} else {
		return wsc.user.GetUsername()
	}
}

func (wsc *WsClient) getClientType() string {
	if wsc.remote != nil {
		return "server"
	} else {
		return "web"
	}
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

	AddSubscription(sub Subscription)

	GetUser() *User
	GetRemote() *Instance

	Error(error)
}

type SocketUser interface {
	SocketType() string
}
