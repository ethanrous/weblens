package client

import (
	"iter"
	"net"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/context"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type ClientId = string

type WsClient struct {
	conn          *websocket.Conn
	user          *user_model.User
	tower         *tower_model.Instance
	connId        ClientId
	subscriptions []websocket_mod.Subscription
	updateMu      sync.Mutex
	subsMu        sync.Mutex
	Active        atomic.Bool

	log zerolog.Logger
}

func NewClient(ctx context.LoggerContext, conn *websocket.Conn, socketUser SocketUser) *WsClient {
	clientId := uuid.New().String()

	newClient := &WsClient{
		connId:   ClientId(clientId),
		conn:     conn,
		updateMu: sync.Mutex{},
		subsMu:   sync.Mutex{},
	}
	newClient.Active.Store(true)

	if socketUser.SocketType() == websocket_mod.WebClient {
		newClient.user = socketUser.(*user_model.User)
	} else if socketUser.SocketType() == websocket_mod.TowerClient {
		newClient.tower = socketUser.(*tower_model.Instance)
	}

	newLogger := ctx.Log().With().Str("client_id", newClient.getClientName()).Str("websocket_id", newClient.GetClientId()).Logger()
	newClient.log = newLogger

	newClient.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("New client connected") })

	return newClient
}

func (wsc *WsClient) IsOpen() bool {
	return wsc.Active.Load()
}

func (wsc *WsClient) GetClientId() ClientId {
	return wsc.connId
}

func (wsc *WsClient) ClientType() websocket_mod.ClientType {
	if wsc.tower != nil {
		return websocket_mod.TowerClient
	}
	return websocket_mod.WebClient
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
		return 0, nil, errors.Errorf("client is closed")
	}
	return wsc.conn.ReadMessage()
}

func (wsc *WsClient) Error(err error) {
	if err != nil {
		wsc.log.Error().Stack().Err(err).Msg("Websocket error")
	}
	err = wsc.Send(websocket_mod.WsResponseInfo{EventTag: websocket_mod.ErrorEvent, Error: err.Error()})
}

func (wsc *WsClient) GetSubscriptions() iter.Seq[websocket_mod.Subscription] {
	wsc.updateMu.Lock()
	defer wsc.updateMu.Unlock()
	return slices.Values(wsc.subscriptions)
}

func (wsc *WsClient) AddSubscription(sub websocket_mod.Subscription) {
	wsc.updateMu.Lock()
	defer wsc.updateMu.Unlock()
	wsc.subscriptions = append(wsc.subscriptions, sub)

	wsc.log.Trace().Func(func(e *zerolog.Event) { e.Str("websocket_subscribe_key", sub.SubscriptionId).Msg("Added Subscription") })
}

func (wsc *WsClient) RemoveSubscription(key string) {
	wsc.updateMu.Lock()
	defer wsc.updateMu.Unlock()

	subIndex := slices.IndexFunc(wsc.subscriptions, func(s websocket_mod.Subscription) bool { return s.SubscriptionId == key })
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

func (wsc *WsClient) Send(msg websocket_mod.WsResponseInfo) error {
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
			return errors.WithStack(err)
		}
	} else {
		return errors.Errorf("trying to send to closed client")
	}

	return nil
}

func (wsc *WsClient) Disconnect() {
	wsc.Active.Store(false)

	wsc.updateMu.Lock()
	defer wsc.updateMu.Unlock()

	err := wsc.conn.Close()
	if err != nil && !errors.Is(err, net.ErrClosed) {
		wsc.log.Error().Stack().Err(err).Msg("")
		return
	}

	wsc.log.Trace().Func(func(e *zerolog.Event) {
		e.Msgf("Disconnected %s client %s [%s]", wsc.getClientType(), wsc.getClientName(), wsc.GetClientId())
	})
}

func (wsc *WsClient) Log() *zerolog.Logger {
	return &wsc.log
}

func (wsc *WsClient) getClientName() string {
	if wsc.tower != nil {
		return wsc.tower.Name
	} else if wsc.user != nil {
		return wsc.user.Username
	}
	return "unknown"
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
	SocketType() websocket_mod.ClientType
}
