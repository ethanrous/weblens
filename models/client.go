package models

import (
	"iter"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/task"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type ClientId = string

var _ Client = (*WsClient)(nil)

type WsClient struct {
	conn          *websocket.Conn
	user          *User
	remote        *Instance
	connId        ClientId
	subscriptions []Subscription
	updateMu      sync.Mutex
	subsMu        sync.Mutex
	Active        atomic.Bool

	log *zerolog.Logger
}

func NewClient(conn *websocket.Conn, socketUser SocketUser, logger *zerolog.Logger) *WsClient {
	clientId := uuid.New().String()

	newClient := &WsClient{
		connId:   ClientId(clientId),
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

	newLogger := logger.With().Str("client_id", newClient.getClientName()).Str("websocket_id", newClient.GetClientId()).Logger()
	newClient.log = &newLogger

	newClient.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("New client connected") })

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
	if wsc.conn == nil || !wsc.Active.Load() {
		return 0, nil, werror.Errorf("client is closed")
	}
	return wsc.conn.ReadMessage()
}

func (wsc *WsClient) Error(err error) {

	safe, _ := werror.GetSafeErr(err)
	err = wsc.Send(WsResponseInfo{EventTag: ErrorEvent, Error: safe.Error()})
	wsc.log.Error().Stack().Err(err).Msg("")
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

	err := wsc.Send(msg)
	if err != nil {
		wsc.log.Error().Stack().Err(err).Msg("Failed to send event")
	}
}

func (wsc *WsClient) PushFileUpdate(updatedFile *fileTree.WeblensFileImpl, media *Media) {
	msg := WsResponseInfo{
		EventTag:      FileUpdatedEvent,
		SubscribeKey:  updatedFile.ID(),
		Content:       WsC{"fileInfo": updatedFile, "mediaData": media},
		BroadcastType: FolderSubscribe,
		SentTime:      time.Now().UnixMilli(),
	}

	err := wsc.Send(msg)
	if err != nil {
		wsc.log.Error().Stack().Err(err).Msg("Failed to send file update")
	}
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

	err := wsc.Send(msg)
	if err != nil {
		wsc.log.Error().Stack().Err(err).Msg("")
	}
}

func (wsc *WsClient) PushPoolUpdate(pool task.Pool, event string, result task.TaskResult) {
	if pool.IsGlobal() {
		wsc.log.Warn().Msg("Not pushing update on global pool")
		return
	}

	msg := WsResponseInfo{
		EventTag:      event,
		SubscribeKey:  pool.ID(),
		Content:       result.ToMap(),
		TaskType:      pool.CreatedInTask().JobName(),
		BroadcastType: TaskSubscribe,
	}

	err := wsc.Send(msg)
	wsc.log.Error().Stack().Err(err).Msg("")
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

	wsc.log.Debug().Func(func(e *zerolog.Event) { e.Str("websocket_subscribe_key", sub.Key).Msg("Added Subscription") })
}

func (wsc *WsClient) RemoveSubscription(key SubId) {
	wsc.updateMu.Lock()
	defer wsc.updateMu.Unlock()

	subIndex := slices.IndexFunc(wsc.subscriptions, func(s Subscription) bool { return s.Key == key })
	if subIndex == -1 {
		return
	}
	wsc.subscriptions, _ = internal.Yoink(wsc.subscriptions, subIndex)

	wsc.log.Debug().Func(func(e *zerolog.Event) { e.Str("websocket_subscribe_key", key).Msg("Removed Subscription") })
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
	if msg.SentTime == 0 {
		msg.SentTime = time.Now().UnixMilli()
	}

	if wsc != nil && wsc.Active.Load() {
		wsc.updateMu.Lock()
		defer wsc.updateMu.Unlock()

		wsc.log.Trace().Func(func(e *zerolog.Event) { e.Str("websocket_event", msg.EventTag).Msg("Sending websocket message") })

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
		wsc.log.Error().Stack().Err(err).Msg("")
		return
	}
	wsc.updateMu.Unlock()

	wsc.log.Trace().Func(func(e *zerolog.Event) {
		e.Msgf("Disconnected %s client [%s]", wsc.getClientType(), wsc.getClientName())
	})
}

func (wsc *WsClient) Log() *zerolog.Logger {
	return wsc.log
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
