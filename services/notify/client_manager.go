package notify

import (
	"context"
	"maps"
	"slices"
	"sync"
	"time"

	client_model "github.com/ethanrous/weblens/models/client"
	websocket_model "github.com/ethanrous/weblens/models/client"
	share_model "github.com/ethanrous/weblens/models/share"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/log"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const notificationChanCapacity = 1000

// ErrSubscriptionNotFound is returned when attempting to unsubscribe from a subscription that does not exist.
var ErrSubscriptionNotFound = wlerrors.New("subscription not found")

// ClientManager manages websocket client connections and their subscriptions to various resources.
type ClientManager struct {
	webClientMap    map[string]*websocket_model.WsClient
	remoteClientMap map[string]*websocket_model.WsClient
	clientMu        sync.RWMutex

	cores   map[string]*websocket_model.WsClient
	coresMu sync.RWMutex

	// Key: subscription identifier, value: clientConn instance
	// Use map to take advantage of O(1) lookup time when finding or removing clients
	// by subscription identifier
	// {
	// 	"fileID": [
	// 		*client1,
	// 		*client2,
	//     ]
	// }
	folderSubs map[string][]*websocket_model.WsClient
	folderMu   sync.Mutex

	taskSubs map[string][]*websocket_model.WsClient
	taskMu   sync.Mutex

	taskTypeSubs map[string][]*websocket_model.WsClient
	taskTypeMu   sync.Mutex

	notificationChan chan websocket_mod.WsResponseInfo
}

// NewClientManager creates and initializes a new ClientManager with a background notification worker.
func NewClientManager(ctx context.Context) *ClientManager {
	cm := &ClientManager{
		webClientMap:    map[string]*websocket_model.WsClient{},
		remoteClientMap: map[string]*websocket_model.WsClient{},
		cores:           map[string]*websocket_model.WsClient{},

		folderSubs:   map[string][]*websocket_model.WsClient{},
		taskSubs:     map[string][]*websocket_model.WsClient{},
		taskTypeSubs: map[string][]*websocket_model.WsClient{},

		notificationChan: make(chan websocket_mod.WsResponseInfo, notificationChanCapacity),
	}

	go cm.notificationWorker(ctx)

	return cm
}

// ClientConnect establishes a new websocket connection for a user and registers the client with the manager.
func (cm *ClientManager) ClientConnect(ctx context.Context, conn *websocket.Conn, user *user_model.User) (*websocket_model.WsClient, error) {
	if user == nil {
		return nil, wlerrors.New("user is nil")
	}

	newClient := websocket_model.NewClient(ctx, conn, user)

	cm.clientMu.Lock()
	cm.webClientMap[newClient.GetClientID()] = newClient
	cm.clientMu.Unlock()

	log.FromContext(ctx).Debug().Msgf("Client [%s] connected", user.Username)

	return newClient, nil
}

// RemoteConnect establishes a websocket connection for a remote instance and notifies all clients of the connection.
func (cm *ClientManager) RemoteConnect(ctx context.Context, conn *websocket.Conn, remote *tower_model.Instance) *websocket_model.WsClient {
	newClient := websocket_model.NewClient(ctx, conn, remote)

	cm.clientMu.Lock()
	cm.remoteClientMap[remote.TowerID] = newClient
	cm.clientMu.Unlock()

	notif := NewSystemNotification(websocket_mod.RemoteConnectionChangedEvent, websocket_mod.WsData{"towerID": remote.TowerID, "online": true})
	cm.Notify(ctx, notif)

	return newClient
}

// ClientDisconnect removes a client from all subscriptions and disconnects them from the manager.
func (cm *ClientManager) ClientDisconnect(ctx context.Context, c *websocket_model.WsClient) error {
	if !c.Active.Load() {
		return wlerrors.New("client is already disconnected")
	}

	for s := range c.GetSubscriptions() {
		err := cm.removeSubscription(ctx, s, c, true)

		// Client is leaving anyway, no point returning an error from here
		// just log it
		if err != nil {
			log.FromContext(ctx).Error().Stack().Err(err).Msg("")
		}
	}

	err := cm.removeClient(ctx, c)
	if err != nil {
		return wlerrors.Errorf("failed to remove client [%s]: %w", c.GetClientID(), err)
	}

	err = c.Disconnect()
	if err != nil {
		return err
	}

	return nil
}

// DisconnectAll disconnects all connected clients from the manager.
func (cm *ClientManager) DisconnectAll(ctx context.Context) error {
	for _, c := range cm.GetAllClients() {
		_ = cm.ClientDisconnect(ctx, c)
	}

	return nil
}

// Notify queues one or more websocket messages to be sent to clients by the notification worker.
func (cm *ClientManager) Notify(ctx context.Context, msg ...websocket_mod.WsResponseInfo) {
	select {
	case <-ctx.Done():
		log.FromContext(ctx).Error().Stack().Err(wlerrors.WithStack(ctx.Err())).Msgf("Context done, not sending websocket message: %s", msg[0].EventTag)

		return
	default:
	}

	for _, m := range msg {
		cm.notificationChan <- m
	}
}

// Flush loads a no-op message into the notification channel as a sort of "tracer round", and then waits for the notification worker to process it.
// This is useful to ensure that all pending notifications are sent before a task forces all clients to unsubscribe, etc.
func (cm *ClientManager) Flush(ctx context.Context) {
	if ctx.Err() != nil {
		log.FromContext(ctx).Error().Stack().Err(wlerrors.WithStack(ctx.Err())).Msg("Context is done, not flushing notifications")

		return
	}

	done := make(chan struct{})
	cm.notificationChan <- websocket_mod.WsResponseInfo{
		EventTag: websocket_mod.FlushEvent,
		Sent:     done,
	}

	select {
	case <-done:
		log.FromContext(ctx).Trace().Msg("Flushed notifications")
	case <-time.After(5 * time.Second):
		log.FromContext(ctx).Warn().Msg("Flush timed out")
	}
}

// GetClientByUsername returns the websocket client for the specified username, or nil if not found.
func (cm *ClientManager) GetClientByUsername(username string) *websocket_model.WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()

	// FIXME: This is O(n) and should be better
	for _, c := range cm.webClientMap {
		if c.GetUser().GetUsername() == username {
			return c
		}
	}

	return nil
}

// GetClientByTowerID returns the websocket client for the specified tower ID, or nil if not found.
func (cm *ClientManager) GetClientByTowerID(towerID string) *websocket_model.WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()

	return cm.remoteClientMap[towerID]
}

// GetCoreByTowerID returns the core websocket client for the specified tower ID, or nil if not found.
func (cm *ClientManager) GetCoreByTowerID(towerID string) *websocket_model.WsClient {
	cm.coresMu.RLock()
	defer cm.coresMu.RUnlock()

	return cm.cores[towerID]
}

// GetAllClients returns a slice of all connected web and remote clients.
func (cm *ClientManager) GetAllClients() []*websocket_model.WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()

	return append(slices.Collect(maps.Values(cm.webClientMap)), slices.Collect(maps.Values(cm.remoteClientMap))...)
}

// GetConnectedAdmins returns a slice of all connected clients that have admin privileges.
func (cm *ClientManager) GetConnectedAdmins() []*websocket_model.WsClient {
	clients := cm.GetAllClients()
	admins := slices_mod.Filter(
		clients, func(c *websocket_model.WsClient) bool {
			return c.GetUser().IsAdmin()
		},
	)

	return admins
}

// GetSubscribers returns a slice of all clients subscribed to the specified subscription type and key.
func (cm *ClientManager) GetSubscribers(ctx context.Context, st websocket_mod.SubscriptionType, key string) (clients []*websocket_model.WsClient) {
	switch st {
	case websocket_mod.FolderSubscribe:
		{
			cm.folderMu.Lock()
			clients = cm.folderSubs[key]
			cm.folderMu.Unlock()
		}
	case websocket_mod.TaskSubscribe:
		{
			cm.taskMu.Lock()
			clients = cm.taskSubs[key]
			cm.taskMu.Unlock()
		}
	case websocket_mod.UserSubscribe:
		{
			cm.clientMu.Lock()
			allClients := slices.Collect(maps.Values(cm.webClientMap))
			cm.clientMu.Unlock()

			clients = slices_mod.Filter(
				allClients, func(c *websocket_model.WsClient) bool {
					return c.GetUser().GetUsername() == key
				},
			)
		}
	case websocket_mod.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			clients = cm.taskTypeSubs[key]
			cm.taskTypeMu.Unlock()
		}
	default:
		err := wlerrors.Errorf("Unknown subscriber type: [%s]", st)
		log.FromContext(ctx).Error().Stack().Err(err).Msgf("Failed to get subscribers for key [%s]", key)
	}

	for c := range slices.Values(clients) {
		if c != nil {
			continue
		}

		log.FromContext(ctx).Error().Msgf("Client is nil!")
	}

	// Copy clients to not modify reference in the map
	return slices.Clone(clients)
}

// SubscribeToFile subscribes a client to receive notifications for changes to a specific file.
func (cm *ClientManager) SubscribeToFile(ctx context.Context, c *client_model.WsClient, file IDer, _ *share_model.FileShare, subTime time.Time) error {
	if file == nil {
		return wlerrors.New("file is nil")
	}

	// TODO: check share

	sub := websocket_mod.Subscription{Type: websocket_mod.FolderSubscribe, SubscriptionID: file.ID(), When: subTime}
	cm.addSubscription(ctx, sub, c)

	return nil
}

// SubscribeToTask subscribes a client to receive notifications for a specific task.
func (cm *ClientManager) SubscribeToTask(ctx context.Context, c *client_model.WsClient, task IDer, subTime time.Time) error {
	// if done, _ := task.Status(); done {
	// 	log.FromContext(ctx).Debug().Msgf("Task [%s] is already done, not subscribing", task.ID())
	//
	// 	return task_model.ErrTaskAlreadyComplete
	// }
	sub := websocket_mod.Subscription{Type: websocket_mod.TaskSubscribe, SubscriptionID: task.ID(), When: subTime}

	cm.addSubscription(ctx, sub, c)

	// notif := NewTaskNotification(task, websocket_mod.TaskCreatedEvent, task.GetMeta().FormatToResult())

	// err := c.Send(notif)
	// if err != nil {
	// 	return err
	// }

	return nil
}

// Unsubscribe removes a client's subscription to the specified key if the subscription exists and was created before the unsubscribe time.
func (cm *ClientManager) Unsubscribe(ctx context.Context, client *websocket_model.WsClient, key string, unSubTime time.Time) error {
	client.SubLock()
	defer client.SubUnlock()

	var targetSub websocket_mod.Subscription

	for sub := range client.GetSubscriptions() {
		if sub.SubscriptionID != key {
			continue
		}

		if !sub.When.Before(unSubTime) {
			log.FromContext(ctx).Debug().Func(func(e *zerolog.Event) { e.Msgf("Ignoring unsubscribe request that happened before subscribe request") })

			continue
		}

		targetSub = sub

		break
	}

	if targetSub == (websocket_mod.Subscription{}) {
		return wlerrors.WithStack(ErrSubscriptionNotFound)
	}

	client.RemoveSubscription(key)

	return cm.removeSubscription(ctx, targetSub, client, false)
}

// FolderSubToTask subscribes all clients watching a folder to a task associated with that folder.
func (cm *ClientManager) FolderSubToTask(ctx context.Context, folderID string, task IDer) {
	subs := cm.GetSubscribers(ctx, websocket_mod.FolderSubscribe, folderID)

	subTime := time.Now()
	for _, s := range subs {
		err := cm.SubscribeToTask(ctx, s, task, subTime)
		if err != nil {
			log.FromContext(ctx).Error().Stack().Err(err).Msg("")
		}
	}
}

// UnsubscribeAllByID unsubscribes all clients from a subscription, by its ID. Useful to unsub all from a task when it completes or is canceled, for example.
func (cm *ClientManager) UnsubscribeAllByID(ctx context.Context, subID string, subscriptionType websocket_mod.SubscriptionType) error {
	subs := cm.GetSubscribers(ctx, subscriptionType, subID)

	for _, s := range subs {
		err := cm.Unsubscribe(ctx, s, subID, time.Now())
		if err != nil && !wlerrors.Is(err, ErrSubscriptionNotFound) {
			return err
		} else if err != nil {
			log.FromContext(ctx).Warn().Msgf("Subscription [%s] not found in UnsubscribeAllByID", subID)

			continue
		}
	}

	return nil
}

// Send broadcasts a websocket message to all clients subscribed to the message's subscription key.
func (cm *ClientManager) Send(ctx context.Context, msg websocket_mod.WsResponseInfo) {
	defer wsRecover(ctx)

	if msg.SubscribeKey == "" {
		log.FromContext(ctx).Error().Stack().Err(wlerrors.New("trying to broadcast on empty key")).Msg("Failed to send websocket message")

		return
	}

	var clients []*websocket_model.WsClient

	if msg.BroadcastType == websocket_mod.SystemSubscribe {
		clients = cm.GetAllClients()
	} else {
		clients = cm.GetSubscribers(ctx, msg.BroadcastType, msg.SubscribeKey)
		clients = slices_mod.OnlyUnique(clients)

		if msg.BroadcastType == websocket_mod.TaskSubscribe {
			clients = append(
				clients, cm.GetSubscribers(
					ctx,
					websocket_mod.TaskTypeSubscribe,
					msg.TaskType,
				)...,
			)
		}
	}

	if msg.EventTag == "taskCanceled" {
		log.FromContext(ctx).Debug().Msgf("websocket_event: %d clients for task %s - %s", len(clients), msg.SubscribeKey, msg.BroadcastType)
	}

	// Don't relay messages to the client that sent them
	if msg.RelaySource != "" {
		i := slices.IndexFunc(clients, func(c *websocket_model.WsClient) bool {
			if c.ClientType() == websocket_mod.TowerClient {
				return c.GetInstance().TowerID == msg.RelaySource
			} else if c.ClientType() == websocket_mod.WebClient {
				return c.GetUser().GetUsername() == msg.RelaySource
			}

			return false
		})
		if i != -1 {
			clients = slices.Delete(clients, i, i+1)
		}
	}

	if len(clients) != 0 {
		log.FromContext(ctx).Trace().Str("websocket_event", string(msg.EventTag)).Msgf("Sending [%s] websocket message to %d client(s)", msg.EventTag, len(clients))

		for _, c := range clients {
			err := c.Send(msg)
			if err != nil {
				log.FromContext(ctx).Error().Stack().Err(err).Msg("")
			}
		}
	} else {
		log.FromContext(ctx).Trace().Func(func(e *zerolog.Event) {
			e.Msgf("No subscribers to [%s]. Trying to send [%s]", msg.SubscribeKey, msg.EventTag)
		})

		return
	}
}

// Relay forwards a websocket message to connected remote instances.
func (cm *ClientManager) Relay(_ websocket_mod.WsResponseInfo) {
	panic("not implemented")
}

func (cm *ClientManager) addSubscription(ctx context.Context, subInfo websocket_mod.Subscription, client *websocket_model.WsClient) {
	switch subInfo.Type {
	case websocket_mod.FolderSubscribe:
		{
			cm.folderMu.Lock()
			addSub(cm.folderSubs, subInfo, client)
			cm.folderMu.Unlock()

			log.FromContext(ctx).Trace().Msgf("Added folder subscription [%s] for client [%s]", subInfo.SubscriptionID, client.GetClientID())
		}
	case websocket_mod.TaskSubscribe:
		{
			cm.taskMu.Lock()
			addSub(cm.taskSubs, subInfo, client)
			cm.taskMu.Unlock()

			log.FromContext(ctx).Trace().Msgf("Added task subscription [%s] for client [%s]", subInfo.SubscriptionID, client.GetClientID())
		}
	case websocket_mod.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			addSub(cm.taskTypeSubs, subInfo, client)
			cm.taskTypeMu.Unlock()

			log.FromContext(ctx).Trace().Msgf("Added task type subscription [%s] for client [%s]", subInfo.SubscriptionID, client.GetClientID())
		}
	default:
		{
			log.FromContext(ctx).Error().Msgf("Unknown subType: %s", subInfo.Type)

			return
		}
	}

	client.AddSubscription(subInfo)
}

func (cm *ClientManager) removeSubscription(
	ctx context.Context, subInfo websocket_mod.Subscription, client *websocket_model.WsClient, removeAll bool,
) error {
	var err error

	switch subInfo.Type {
	case websocket_mod.FolderSubscribe:
		{
			cm.folderMu.Lock()
			err = removeSubs(cm.folderSubs, subInfo, client, removeAll)
			cm.folderMu.Unlock()

			log.FromContext(ctx).Trace().Msgf("Removed file subscription [%s] for client [%s]", subInfo.SubscriptionID, client.GetClientID())
		}
	case websocket_mod.TaskSubscribe:
		{
			cm.taskMu.Lock()
			err = removeSubs(cm.taskSubs, subInfo, client, removeAll)
			cm.taskMu.Unlock()

			log.FromContext(ctx).Trace().Msgf("Removed task subscription [%s] for client [%s]", subInfo.SubscriptionID, client.GetClientID())
		}
	case websocket_mod.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			err = removeSubs(cm.taskTypeSubs, subInfo, client, removeAll)
			cm.taskTypeMu.Unlock()
		}
	default:
		{
			return wlerrors.Errorf("Trying to remove unknown subscription type [%s]", subInfo.Type)
		}
	}

	return err
}

func (cm *ClientManager) removeClient(ctx context.Context, client *websocket_model.WsClient) error {
	if client.GetUser() != nil {
		cm.clientMu.Lock()
		delete(cm.webClientMap, client.GetClientID())
		cm.clientMu.Unlock()

		log.FromContext(ctx).Debug().Msgf("Client [%s] disconnected", client.GetUser().GetUsername())
	} else if remote := client.GetInstance(); remote != nil {
		cm.clientMu.Lock()
		delete(cm.remoteClientMap, client.GetInstance().TowerID)
		cm.clientMu.Unlock()

		notif := NewSystemNotification(websocket_mod.RemoteConnectionChangedEvent, websocket_mod.WsData{"towerID": remote.TowerID, "online": false})

		cm.Notify(ctx, notif)
	} else {
		return wlerrors.New("client is not a remote or a user")
	}

	return nil
}

func (cm *ClientManager) notificationWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.FromContext(ctx).Debug().Msg("Notification worker stopped")

			return
		case msg := <-cm.notificationChan:
			if msg.EventTag == "" {
				log.FromContext(ctx).Error().Msg("Received empty event tag in notification worker")

				continue
			} else if msg.EventTag == websocket_mod.FlushEvent {
				log.FromContext(ctx).Trace().Msg("Received flush event, closing sent channel")
			} else {
				cm.Send(ctx, msg)
			}

			if msg.Sent != nil {
				close(msg.Sent)
			}
		}
	}
}

func addSub(subMap map[string][]*websocket_model.WsClient, subInfo websocket_mod.Subscription, client *websocket_model.WsClient) {
	subs, ok := subMap[subInfo.SubscriptionID]

	if !ok {
		subs = []*websocket_model.WsClient{}
	}

	subMap[subInfo.SubscriptionID] = append(subs, client)
}

func removeSubs(
	subMap map[string][]*websocket_model.WsClient, subInfo websocket_mod.Subscription, client *websocket_model.WsClient, removeAll bool,
) error {
	subs, ok := subMap[subInfo.SubscriptionID]
	if !ok {
		return wlerrors.Errorf("Tried to unsubscribe from non-existent key [%s]", subInfo.SubscriptionID)
	}

	if removeAll {
		subs = slices_mod.Filter(subs, func(c *websocket_model.WsClient) bool { return c.GetClientID() != client.GetClientID() })
	} else {
		index := slices.IndexFunc(
			subs, func(c *websocket_model.WsClient) bool { return c.GetClientID() == client.GetClientID() },
		)
		if index != -1 {
			subs = slices.Delete(subs, index, index+1)
		}
	}

	subMap[subInfo.SubscriptionID] = subs

	return nil
}

func wsRecover(ctx context.Context) {
	e := recover()
	if e == nil {
		return
	}

	err, ok := e.(error)
	if !ok {
		err = wlerrors.Errorf("%v", e)
	}

	log.FromContext(ctx).Error().Stack().Err(err).Msg("Websocket send panicked")
}

// IDer is an interface for types that have an ID method returning a string.
type IDer interface {
	ID() string
}
