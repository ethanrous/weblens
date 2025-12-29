// Package client manages websocket client connections and their lifecycle.
package client

import (
	"encoding/json"
	"iter"
	"net"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/errors"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// ID is a unique identifier for websocket clients.
type ID = string

// WsClient represents a websocket client connection.
type WsClient struct {
	conn          *websocket.Conn
	user          *user_model.User
	tower         *tower_model.Instance
	connID        ID
	subscriptions []websocket_mod.Subscription
	updateMu      sync.Mutex
	subsMu        sync.Mutex
	Active        atomic.Bool

	log zerolog.Logger
}

const (
	subscribeKeyLogKey            = "websocket_subscribe_key"
	clientIDLogKey                = "client_id"
	websocketIDLogKey             = "websocket_id"
	websocketMessageContentLogKey = "websocket_message_content"
)

// NewClient creates a new websocket client instance.
func NewClient(ctx context.LoggerContext, conn *websocket.Conn, socketUser SocketUser) *WsClient {
	clientID := uuid.New().String()

	newClient := &WsClient{
		connID:   ID(clientID),
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

	newLogger := ctx.Log().With().Str(clientIDLogKey, newClient.getClientName()).Str(websocketIDLogKey, newClient.GetClientID()).Logger()
	newClient.log = newLogger

	newClient.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("New client connected") })

	return newClient
}

// IsOpen returns whether the client connection is currently active.
func (wsc *WsClient) IsOpen() bool {
	return wsc.Active.Load()
}

// GetClientID returns the unique identifier for this client.
func (wsc *WsClient) GetClientID() ID {
	return wsc.connID
}

// ClientType returns the type of this client (web or tower).
func (wsc *WsClient) ClientType() websocket_mod.ClientType {
	if wsc.tower != nil {
		return websocket_mod.TowerClient
	}

	return websocket_mod.WebClient
}

// GetShortID returns a shortened version of the client ID for display purposes.
func (wsc *WsClient) GetShortID() ID {
	if wsc.connID == "" {
		return ""
	}

	return wsc.connID[28:]
}

// GetUser returns the user associated with this client connection.
func (wsc *WsClient) GetUser() *user_model.User {
	return wsc.user
}

// GetInstance returns the tower instance associated with this client connection.
func (wsc *WsClient) GetInstance() *tower_model.Instance {
	return wsc.tower
}

// ReadOne reads a single message from the websocket connection.
func (wsc *WsClient) ReadOne() (int, []byte, error) {
	if wsc.conn == nil || !wsc.Active.Load() {
		return 0, nil, errors.Errorf("client is closed")
	}

	return wsc.conn.ReadMessage()
}

// Error logs an error and sends it to the client.
func (wsc *WsClient) Error(err error) {
	if err != nil {
		wsc.log.Error().Stack().Err(err).Msg("Websocket error")
	}

	err = wsc.Send(websocket_mod.WsResponseInfo{EventTag: websocket_mod.ErrorEvent, Error: err.Error()})
	if err != nil {
		wsc.log.Error().Stack().Err(err).Msg("Failed to send error message")
	}
}

// GetSubscriptions returns an iterator over the client's current subscriptions.
func (wsc *WsClient) GetSubscriptions() iter.Seq[websocket_mod.Subscription] {
	wsc.updateMu.Lock()
	defer wsc.updateMu.Unlock()

	return slices.Values(wsc.subscriptions)
}

// AddSubscription adds a new subscription to the client.
func (wsc *WsClient) AddSubscription(sub websocket_mod.Subscription) {
	wsc.updateMu.Lock()
	defer wsc.updateMu.Unlock()

	wsc.subscriptions = append(wsc.subscriptions, sub)
}

// RemoveSubscription removes a subscription from the client by key.
func (wsc *WsClient) RemoveSubscription(key string) {
	wsc.updateMu.Lock()
	defer wsc.updateMu.Unlock()

	subIndex := slices.IndexFunc(wsc.subscriptions, func(s websocket_mod.Subscription) bool { return s.SubscriptionID == key })
	if subIndex == -1 {
		return
	}

	wsc.subscriptions = slices.Delete(wsc.subscriptions, subIndex, subIndex+1)
}

// Raw sends a raw message to the client without additional formatting.
func (wsc *WsClient) Raw(msg any) error {
	return wsc.conn.WriteJSON(msg)
}

// SubLock acquires the subscription mutex lock.
func (wsc *WsClient) SubLock() {
	wsc.subsMu.Lock()
}

// SubUnlock releases the subscription mutex lock.
func (wsc *WsClient) SubUnlock() {
	wsc.subsMu.Unlock()
}

// Send sends a websocket message to the client.
func (wsc *WsClient) Send(msg websocket_mod.WsResponseInfo) error {
	if msg.SentTime == 0 {
		msg.SentTime = time.Now().UnixMilli()
	}

	if wsc != nil && wsc.Active.Load() {
		wsc.updateMu.Lock()
		defer wsc.updateMu.Unlock()

		wsc.log.Trace().Func(func(e *zerolog.Event) {
			msgbs, err := json.Marshal(msg)
			if err != nil {
				return
			}

			e.Str("websocket_event", string(msg.EventTag)).Str(websocketMessageContentLogKey, string(msgbs)).Msg("Sending websocket message")
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

// Disconnect closes the websocket connection and marks the client as inactive.
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
		e.Msgf("Disconnected %s client %s [%s]", wsc.getClientType(), wsc.getClientName(), wsc.GetClientID())
	})
}

// Log returns the logger instance for this client.
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
	}

	return "web"
}

// SocketUser represents a user that can connect via websocket.
type SocketUser interface {
	SocketType() websocket_mod.ClientType
}
