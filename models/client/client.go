package client

import (
	"iter"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/internal/werror"
	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/task"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type ClientId = string

type WsClient struct {
	conn          *websocket.Conn
	user          *user_model.User
	tower        *tower_model.Instance
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
		newClient.user = socketUser.(*user_model.User)
	} else if socketUser.SocketType() == "serverClient" {
		newClient.tower = socketUser.(*tower_model.Instance)
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
	if wsc.tower != nil {
		return InstanceClient
	}
	return WebClient
}

func (wsc *WsClient) GetShortId() ClientId {
	if wsc.connId == "" {
		return ""
	}
	return wsc.connId[28:]
}

func (wsc *WsClient) GetUser() *user_model.User {
	return wsc.user
}

func (wsc *WsClient) GetInstance() *tower_model.Instance {
	return wsc.tower
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
	if err != nil {
		wsc.log.Error().Stack().Err(err).Msg("")
	}
}

func (wsc *WsClient) PushWeblensEvent(eventTag WsEvent, content ...WsC) {
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

func (wsc *WsClient) PushFileUpdate(updatedFile *file_model.WeblensFileImpl, media *media_model.Media) {
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

func (wsc *WsClient) PushTaskUpdate(task *task.Task, event WsEvent, result task.TaskResult) {
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

func (wsc *WsClient) PushPoolUpdate(pool task.Pool, event WsEvent, result task.TaskResult) {
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

	wsc.log.Debug().Func(func(e *zerolog.Event) { e.Str("websocket_subscribe_key", sub.SubscriptionId).Msg("Added Subscription") })
}

func (wsc *WsClient) RemoveSubscription(key string) {
	wsc.updateMu.Lock()
	defer wsc.updateMu.Unlock()

	subIndex := slices.IndexFunc(wsc.subscriptions, func(s Subscription) bool { return s.SubscriptionId == key })
	if subIndex == -1 {
		return
	}
	wsc.subscriptions = slices.Delete(wsc.subscriptions, subIndex, subIndex+1)

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

		wsc.log.Trace().Func(func(e *zerolog.Event) {
			e.Str("websocket_event", string(msg.EventTag)).Msg("Sending websocket message")
		})

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
	if wsc.tower != nil {
		return wsc.tower.Name
	} else {
		return wsc.user.Username
	}
}

func (wsc *WsClient) getClientType() string {
	if wsc.tower != nil {
		return "server"
	} else {
		return "web"
	}
}

// type Client interface {
// 	BasicCaster
//
// 	PushFileUpdate(updatedFile *file_model.WeblensFile, media *media_model.Media)
//
// 	IsOpen() bool
//
// 	ReadOne() (int, []byte, error)
//
// 	GetSubscriptions() iter.Seq[Subscription]
// 	GetClientId() ClientId
// 	GetShortId() ClientId
//
// 	SubLock()
// 	SubUnlock()
//
// 	AddSubscription(sub Subscription)
//
// 	GetUser() *user_model.User
// 	GetInstance() *tower_model.Instance
//
// 	Error(error)
// }

type SocketUser interface {
	SocketType() string
}
