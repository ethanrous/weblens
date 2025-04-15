package notify

import (
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
	"github.com/ethanrous/weblens/modules/context"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	"github.com/ethanrous/weblens/modules/structs"
	task_mod "github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var ErrSubscriptionNotFound = errors.New("subscription not found")

var _ client_model.ClientManager = &ClientManager{}

type ClientManager struct {
	webClientMap    map[string]*websocket_model.WsClient
	remoteClientMap map[string]*websocket_model.WsClient

	core *websocket_model.WsClient

	// Key: subscription identifier, value: clientConn instance
	// Use map to take advantage of O(1) lookup time when finding or removing clients
	// by subscription identifier
	// {
	// 	"fileId": [
	// 		*client1,
	// 		*client2,
	//     ]
	// }
	folderSubs   map[string][]*websocket_model.WsClient
	taskSubs     map[string][]*websocket_model.WsClient
	taskTypeSubs map[string][]*websocket_model.WsClient
	clientMu     sync.RWMutex
	folderMu     sync.Mutex
	taskMu       sync.Mutex
	taskTypeMu   sync.Mutex
}

func NewClientManager() *ClientManager {
	cm := &ClientManager{
		webClientMap:    map[string]*websocket_model.WsClient{},
		remoteClientMap: map[string]*websocket_model.WsClient{},

		folderSubs:   map[string][]*websocket_model.WsClient{},
		taskSubs:     map[string][]*websocket_model.WsClient{},
		taskTypeSubs: map[string][]*websocket_model.WsClient{},
	}

	return cm
}

func (cm *ClientManager) ClientConnect(ctx context.LoggerContext, conn *websocket.Conn, user *user_model.User) (*websocket_model.WsClient, error) {
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

func (cm *ClientManager) RemoteConnect(ctx context.LoggerContext, conn *websocket.Conn, remote *tower_model.Instance) *websocket_model.WsClient {
	newClient := websocket_model.NewClient(ctx, conn, remote)

	cm.clientMu.Lock()
	cm.remoteClientMap[remote.TowerId] = newClient
	cm.clientMu.Unlock()

	if remote.IsCore() {
		cm.core = newClient
	}

	notif := NewSystemNotification(websocket_mod.RemoteConnectionChangedEvent, websocket_mod.WsData{"remoteId": remote.TowerId, "online": true})
	cm.Notify(notif)

	return newClient
}

func (cm *ClientManager) ClientDisconnect(ctx context.LoggerContext, c *websocket_model.WsClient) {
	for s := range c.GetSubscriptions() {
		err := cm.removeSubscription(s, c, true)

		// Client is leaving anyway, no point returning an error from here
		// just log it
		if err != nil {
			ctx.Log().Error().Stack().Err(err).Msg("")
		}
	}

	cm.clientMu.Lock()
	defer cm.clientMu.Unlock()

	if c.GetUser() != nil {
		delete(cm.webClientMap, c.GetClientId())
		ctx.Log().Debug().Msgf("Client [%s] disconnected", c.GetUser().GetUsername())
	} else if remote := c.GetInstance(); remote != nil {
		delete(cm.remoteClientMap, c.GetInstance().TowerId)
		notif := NewSystemNotification(websocket_mod.RemoteConnectionChangedEvent, websocket_mod.WsData{"remoteId": remote.TowerId, "online": false})

		cm.Notify(notif)
	}

	c.Disconnect()
}

func (cm *ClientManager) Notify(msg ...websocket_mod.WsResponseInfo) {
	for _, m := range msg {
		cm.Send(context.Background(), m)
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

func (cm *ClientManager) GetClientByServerId(instanceId string) *websocket_model.WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return cm.remoteClientMap[instanceId]
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

func (cm *ClientManager) GetSubscribers(ctx context.LoggerContext, st websocket_mod.WsAction, key string) (clients []*websocket_model.WsClient) {
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
		ctx.Log().Error().Msgf("Unknown subscriber type: [%s]", st)
	}

	// Copy clients to not modify reference in the map
	return clients[:]
}

func (cm *ClientManager) SubscribeToFile(ctx context.ContextZ, c *client_model.WsClient, file *file_model.WeblensFileImpl, share *share_model.FileShare, subTime time.Time) error {
	if file == nil {
		return errors.New("file is nil")
	}

	sub := websocket_mod.Subscription{Type: websocket_mod.FolderSubscribe, SubscriptionId: file.ID(), When: subTime}
	cm.addSubscription(ctx, sub, c)

	notifs := NewFileNotification(ctx, file, websocket_mod.FileUpdatedEvent, structs.MediaInfo{})
	cm.Notify(notifs...)

	return nil
}

func (cm *ClientManager) SubscribeToTask(ctx context.LoggerContext, c *client_model.WsClient, task *task_model.Task, subTime time.Time) error {
	if done, _ := task.Status(); done {
		return nil
	}

	sub := websocket_mod.Subscription{Type: websocket_mod.TaskSubscribe, SubscriptionId: task.Id(), When: subTime}

	cm.addSubscription(ctx, sub, c)

	notif := NewTaskNotification(task, websocket_mod.TaskCreatedEvent, task.GetMeta().FormatToResult())
	cm.Notify(notif)

	return nil
}

func (cm *ClientManager) Unsubscribe(ctx context.LoggerContext, c *websocket_model.WsClient, key string, unSubTime time.Time) error {
	c.SubLock()
	defer c.SubUnlock()

	var sub websocket_mod.Subscription
	for s := range c.GetSubscriptions() {
		if s.SubscriptionId == key && !s.When.Before(unSubTime) {
			ctx.Log().Debug().Func(func(e *zerolog.Event) { e.Msgf("Ignoring unsubscribe request that happened before subscribe request") })
			continue
		}

		if s.SubscriptionId == key && s.When.Before(unSubTime) {
			sub = s
			break
		}
	}

	if sub == (websocket_mod.Subscription{}) {
		return errors.WithStack(ErrSubscriptionNotFound)
	}

	c.RemoveSubscription(key)
	return cm.removeSubscription(sub, c, false)
}

func (cm *ClientManager) FolderSubToTask(ctx context.LoggerContext, folderId string, task task_mod.Task) {
	subs := cm.GetSubscribers(ctx, websocket_mod.FolderSubscribe, folderId)

	subTime := time.Now()
	for _, s := range subs {
		err := cm.SubscribeToTask(ctx, s, task.(*task_model.Task), subTime)
		if err != nil {
			ctx.Log().Error().Stack().Err(err).Msg("")
		}
	}
}

func (cm *ClientManager) UnsubTask(ctx context.LoggerContext, taskId string) {
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

func wsRecover(ctx context.LoggerContext) {
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

func (cm *ClientManager) Send(ctx context.LoggerContext, msg websocket_mod.WsResponseInfo) {
	defer wsRecover(ctx)

	if msg.SubscribeKey == "" {
		ctx.Log().Error().Stack().Err(errors.New("trying to broadcast on empty key")).Msg("Failed to send websocket message")
		return
	}

	var clients []*websocket_model.WsClient

	if msg.BroadcastType == "serverEvent" {
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

}

// func (cm *ClientManager) PushWeblensEvent(event websocket_mod.WsEvent, msg websocket_mod.WsData) {
// 	cm.Send(websocket_mod.WsResponseInfo{
// 		EventTag:      event,
// 		Content:       msg,
// 		BroadcastType: "serverEvent",
// 	})
// }

func (cm *ClientManager) addSubscription(ctx context.LoggerContext, subInfo websocket_mod.Subscription, client *websocket_model.WsClient) {
	switch subInfo.Type {
	case websocket_mod.FolderSubscribe:
		{
			cm.folderMu.Lock()
			addSub(cm.folderSubs, subInfo, client)
			cm.folderMu.Unlock()
		}
	case websocket_mod.TaskSubscribe:
		{
			cm.taskMu.Lock()
			addSub(cm.taskSubs, subInfo, client)
			cm.taskMu.Unlock()
		}
	case websocket_mod.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			addSub(cm.taskTypeSubs, subInfo, client)
			cm.taskTypeMu.Unlock()
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
