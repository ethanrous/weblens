package notify

import (
	"context"
	"maps"
	"slices"
	"sync"
	"time"

	client_model "github.com/ethanrous/weblens/models/client"
	websocket_model "github.com/ethanrous/weblens/models/client"
	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	task_model "github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	task_mod "github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const notificationChanCapacity = 1000

var ErrSubscriptionNotFound = errors.New("subscription not found")

var _ client_model.ClientManager = &ClientManager{}

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
	// 	"fileId": [
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

func (cm *ClientManager) ClientConnect(ctx context_mod.LoggerContext, conn *websocket.Conn, user *user_model.User) (*websocket_model.WsClient, error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}

	newClient := websocket_model.NewClient(ctx, conn, user)

	cm.clientMu.Lock()
	cm.webClientMap[newClient.GetClientId()] = newClient
	cm.clientMu.Unlock()

	ctx.Log().Debug().Msgf("Client [%s] connected", user.Username)

	return newClient, nil
}

func (cm *ClientManager) RemoteConnect(ctx context_mod.LoggerContext, conn *websocket.Conn, remote *tower_model.Instance) *websocket_model.WsClient {
	newClient := websocket_model.NewClient(ctx, conn, remote)

	cm.clientMu.Lock()
	cm.remoteClientMap[remote.TowerId] = newClient
	cm.clientMu.Unlock()

	notif := NewSystemNotification(websocket_mod.RemoteConnectionChangedEvent, websocket_mod.WsData{"towerId": remote.TowerId, "online": true})
	cm.Notify(ctx, notif)

	return newClient
}

func (cm *ClientManager) ClientDisconnect(ctx context.Context, c *websocket_model.WsClient) {
	if !c.Active.Load() {
		return
	}

	for s := range c.GetSubscriptions() {
		err := cm.removeSubscription(s, c, true)

		// Client is leaving anyway, no point returning an error from here
		// just log it
		if err != nil {
			context_mod.ToZ(ctx).Log().Error().Stack().Err(err).Msg("")
		}
	}

	err := cm.removeClient(ctx, c)
	if err != nil {
		context_mod.ToZ(ctx).Log().Error().Stack().Err(err).Msgf("Failed to remove client [%s]", c.GetClientId())
	}

	c.Disconnect()
}

func (cm *ClientManager) DisconnectAll(ctx context.Context) {
	for _, c := range cm.GetAllClients() {
		cm.ClientDisconnect(ctx, c)
	}
}

func (cm *ClientManager) Notify(ctx context.Context, msg ...websocket_mod.WsResponseInfo) {
	select {
	case <-ctx.Done():
		log.FromContext(ctx).Trace().Msgf("Context done, not sending websocket message: %s", msg[0].EventTag)

		return
	default:
	}

	for _, m := range msg {
		cm.notificationChan <- m
	}
}

const flushEventTag = "flush"

// Flush loads a no-op message into the notification channel as a sort of "tracer round", and then waits for the notification worker to process it.
// This is useful to ensure that all pending notifications are sent before a task forces all clients to unsubscribe, etc.
func (cm *ClientManager) Flush(ctx context.Context) {
	if ctx.Err() != nil {
		log.FromContext(ctx).Debug().Msg("Context is done, not flushing notifications")

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

func (cm *ClientManager) GetClientByTowerId(towerId string) *websocket_model.WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()

	return cm.remoteClientMap[towerId]
}

func (cm *ClientManager) GetCoreByTowerId(towerId string) *websocket_model.WsClient {
	cm.coresMu.RLock()
	defer cm.coresMu.RUnlock()

	return cm.cores[towerId]
}

func (cm *ClientManager) GetAllClients() []*websocket_model.WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()

	return append(slices.Collect(maps.Values(cm.webClientMap)), slices.Collect(maps.Values(cm.remoteClientMap))...)
}

func (cm *ClientManager) GetConnectedAdmins() []*websocket_model.WsClient {
	clients := cm.GetAllClients()
	admins := slices_mod.Filter(
		clients, func(c *websocket_model.WsClient) bool {
			return c.GetUser().IsAdmin()
		},
	)

	return admins
}

func (cm *ClientManager) GetSubscribers(ctx context_mod.LoggerContext, st websocket_mod.SubscriptionType, key string) (clients []*websocket_model.WsClient) {
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
		err := errors.Errorf("Unknown subscriber type: [%s]", st)
		ctx.Log().Error().Stack().Err(err).Msgf("Failed to get subscribers for key [%s]", key)
	}

	for c := range slices.Values(clients) {
		if c != nil {
			continue
		}

		ctx.Log().Error().Msgf("Client is nil!")
	}

	// Copy clients to not modify reference in the map
	return slices.Clone(clients)
}

func (cm *ClientManager) SubscribeToFile(ctx context_mod.ContextZ, c *client_model.WsClient, file *file_model.WeblensFileImpl, share *share_model.FileShare, subTime time.Time) error {
	if file == nil {
		return errors.New("file is nil")
	}

	sub := websocket_mod.Subscription{Type: websocket_mod.FolderSubscribe, SubscriptionId: file.ID(), When: subTime}
	cm.addSubscription(ctx, sub, c)

	notifs := NewFileNotification(ctx, file, websocket_mod.FileUpdatedEvent)
	cm.Notify(ctx, notifs...)

	return nil
}

func (cm *ClientManager) SubscribeToTask(ctx context_mod.LoggerContext, c *client_model.WsClient, task *task_model.Task, subTime time.Time) error {
	if done, _ := task.Status(); done {
		ctx.Log().Debug().Msgf("Task [%s] is already done, not subscribing", task.Id())

		return task_model.ErrTaskAlreadyComplete
	}

	sub := websocket_mod.Subscription{Type: websocket_mod.TaskSubscribe, SubscriptionId: task.Id(), When: subTime}

	cm.addSubscription(ctx, sub, c)

	notif := NewTaskNotification(task, websocket_mod.TaskCreatedEvent, task.GetMeta().FormatToResult())

	err := c.Send(notif)
	if err != nil {
		return err
	}

	return nil
}

func (cm *ClientManager) Unsubscribe(ctx context_mod.LoggerContext, client *websocket_model.WsClient, key string, unSubTime time.Time) error {
	client.SubLock()
	defer client.SubUnlock()

	var targetSub websocket_mod.Subscription

	for sub := range client.GetSubscriptions() {
		if sub.SubscriptionId != key {
			continue
		}

		if !sub.When.Before(unSubTime) {
			ctx.Log().Debug().Func(func(e *zerolog.Event) { e.Msgf("Ignoring unsubscribe request that happened before subscribe request") })

			continue
		}

		targetSub = sub

		break
	}

	if targetSub == (websocket_mod.Subscription{}) {
		return errors.WithStack(ErrSubscriptionNotFound)
	}

	client.RemoveSubscription(key)

	return cm.removeSubscription(targetSub, client, false)
}

func (cm *ClientManager) FolderSubToTask(ctx context_mod.LoggerContext, folderId string, task task_mod.Task) {
	subs := cm.GetSubscribers(ctx, websocket_mod.FolderSubscribe, folderId)

	subTime := time.Now()
	for _, s := range subs {
		err := cm.SubscribeToTask(ctx, s, task.(*task_model.Task), subTime)
		if err != nil {
			ctx.Log().Error().Stack().Err(err).Msg("")
		}
	}
}

func (cm *ClientManager) UnsubTask(ctx context_mod.LoggerContext, taskId string) {
	subs := cm.GetSubscribers(ctx, websocket_mod.TaskSubscribe, taskId)

	for _, s := range subs {
		err := cm.Unsubscribe(ctx, s, taskId, time.Now())
		if err != nil && !errors.Is(err, ErrSubscriptionNotFound) {
			ctx.Log().Error().Stack().Err(err).Msg("")
		} else if err != nil {
			ctx.Log().Warn().Msgf("Subscription [%s] not found in unsub task", taskId)
		}
	}
}

func (cm *ClientManager) Send(c context.Context, msg websocket_mod.WsResponseInfo) {
	ctx := context_mod.ToZ(c)

	defer wsRecover(ctx)

	if msg.SubscribeKey == "" {
		ctx.Log().Error().Stack().Err(errors.New("trying to broadcast on empty key")).Msg("Failed to send websocket message")

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
		log.FromContext(c).Debug().Msgf("websocket_event: %d clients for task %s - %s", len(clients), msg.SubscribeKey, msg.BroadcastType)
	}

	// Don't relay messages to the client that sent them
	if msg.RelaySource != "" {
		i := slices.IndexFunc(clients, func(c *websocket_model.WsClient) bool {
			if c.ClientType() == websocket_mod.TowerClient {
				return c.GetInstance().TowerId == msg.RelaySource
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
		ctx.Log().Trace().Str("websocket_event", string(msg.EventTag)).Msgf("Sending [%s] websocket message to %d client(s)", msg.EventTag, len(clients))

		for _, c := range clients {
			err := c.Send(msg)
			if err != nil {
				ctx.Log().Error().Stack().Err(err).Msg("")
			}
		}
	} else {
		ctx.Log().Trace().Func(func(e *zerolog.Event) {
			e.Msgf("No subscribers to [%s]. Trying to send [%s]", msg.SubscribeKey, msg.EventTag)
		})

		return
	}
}

func (cm *ClientManager) Relay(msg websocket_mod.WsResponseInfo) {
	panic("not implemented")
}

func (cm *ClientManager) addSubscription(ctx context_mod.LoggerContext, subInfo websocket_mod.Subscription, client *websocket_model.WsClient) {
	switch subInfo.Type {
	case websocket_mod.FolderSubscribe:
		{
			cm.folderMu.Lock()
			addSub(cm.folderSubs, subInfo, client)
			cm.folderMu.Unlock()

			ctx.Log().Trace().Msgf("Added folder subscription [%s] for client [%s]", subInfo.SubscriptionId, client.GetClientId())
		}
	case websocket_mod.TaskSubscribe:
		{
			cm.taskMu.Lock()
			addSub(cm.taskSubs, subInfo, client)
			cm.taskMu.Unlock()

			ctx.Log().Trace().Msgf("Added task subscription [%s] for client [%s]", subInfo.SubscriptionId, client.GetClientId())
		}
	case websocket_mod.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			addSub(cm.taskTypeSubs, subInfo, client)
			cm.taskTypeMu.Unlock()

			ctx.Log().Trace().Msgf("Added task type subscription [%s] for client [%s]", subInfo.SubscriptionId, client.GetClientId())
		}
	default:
		{
			ctx.Log().Error().Msgf("Unknown subType: %s", subInfo.Type)

			return
		}
	}

	client.AddSubscription(subInfo)
}

func (cm *ClientManager) removeSubscription(
	subInfo websocket_mod.Subscription, client *websocket_model.WsClient, removeAll bool,
) error {
	var err error

	switch subInfo.Type {
	case websocket_mod.FolderSubscribe:
		{
			cm.folderMu.Lock()
			err = removeSubs(cm.folderSubs, subInfo, client, removeAll)
			cm.folderMu.Unlock()
		}
	case websocket_mod.TaskSubscribe:
		{
			cm.taskMu.Lock()
			err = removeSubs(cm.taskSubs, subInfo, client, removeAll)
			cm.taskMu.Unlock()

			log.GlobalLogger().Trace().Msgf("Removed task subscription [%s] for client [%s]", subInfo.SubscriptionId, client.GetClientId())
		}
	case websocket_mod.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			err = removeSubs(cm.taskTypeSubs, subInfo, client, removeAll)
			cm.taskTypeMu.Unlock()
		}
	default:
		{
			return errors.Errorf("Trying to remove unknown subscription type [%s]", subInfo.Type)
		}
	}

	return err
}

func (cm *ClientManager) removeClient(ctx context.Context, client *websocket_model.WsClient) error {
	if client.GetUser() != nil {
		cm.clientMu.Lock()
		delete(cm.webClientMap, client.GetClientId())
		cm.clientMu.Unlock()

		context_mod.ToZ(ctx).Log().Debug().Msgf("Client [%s] disconnected", client.GetUser().GetUsername())
	} else if remote := client.GetInstance(); remote != nil {
		cm.clientMu.Lock()
		delete(cm.remoteClientMap, client.GetInstance().TowerId)
		cm.clientMu.Unlock()

		notif := NewSystemNotification(websocket_mod.RemoteConnectionChangedEvent, websocket_mod.WsData{"towerId": remote.TowerId, "online": false})

		cm.Notify(ctx, notif)
	} else {
		return errors.New("client is not a remote or a user")
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
	subs, ok := subMap[subInfo.SubscriptionId]

	if !ok {
		subs = []*websocket_model.WsClient{}
	}

	subMap[subInfo.SubscriptionId] = append(subs, client)
}

func removeSubs(
	subMap map[string][]*websocket_model.WsClient, subInfo websocket_mod.Subscription, client *websocket_model.WsClient, removeAll bool,
) error {
	subs, ok := subMap[subInfo.SubscriptionId]
	if !ok {
		return errors.Errorf("Tried to unsubscribe from non-existent key [%s]", subInfo.SubscriptionId)
	}

	if removeAll {
		subs = slices_mod.Filter(subs, func(c *websocket_model.WsClient) bool { return c.GetClientId() != client.GetClientId() })
	} else {
		index := slices.IndexFunc(
			subs, func(c *websocket_model.WsClient) bool { return c.GetClientId() == client.GetClientId() },
		)
		if index != -1 {
			subs = slices.Delete(subs, index, index+1)
		}
	}

	subMap[subInfo.SubscriptionId] = subs

	return nil
}

func wsRecover(ctx context_mod.LoggerContext) {
	e := recover()
	if e == nil {
		return
	}

	err, ok := e.(error)
	if !ok {
		err = errors.Errorf("%v", e)
	}

	ctx.Log().Error().Stack().Err(err).Msg("Websocket send panicked")
}
