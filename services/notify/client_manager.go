package notify

import (
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/models/client"
	websocket_model "github.com/ethanrous/weblens/models/client"
	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	task_model "github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/context"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var ErrSubscriptionNotFound = errors.New("subscription not found")

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
	folderSubs map[string][]*websocket_model.WsClient

	taskSubs map[string][]*websocket_model.WsClient

	taskTypeSubs map[string][]*websocket_model.WsClient

	clientMu sync.RWMutex

	folderMu sync.Mutex

	taskMu sync.Mutex

	taskTypeMu sync.Mutex

	log *zerolog.Logger
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

func (cm *ClientManager) ClientConnect(conn *websocket.Conn, user *user_model.User) *websocket_model.WsClient {
	newClient := websocket_model.NewClient(conn, user, cm.log)

	cm.clientMu.Lock()
	cm.webClientMap[newClient.GetClientId()] = newClient
	cm.clientMu.Unlock()

	return newClient
}

func (cm *ClientManager) RemoteConnect(ctx context.ContextZ, conn *websocket.Conn, remote *tower_model.Instance) *websocket_model.WsClient {
	newClient := websocket_model.NewClient(conn, remote, cm.log)

	cm.clientMu.Lock()
	cm.remoteClientMap[remote.TowerId] = newClient
	cm.clientMu.Unlock()

	if remote.IsCore() {
		cm.core = newClient
	}

	cm.PushWeblensEvent(websocket_mod.RemoteConnectionChangedEvent, websocket_mod.WsData{"remoteId": remote.TowerId, "online": true})
	return newClient
}

func (cm *ClientManager) ClientDisconnect(c *websocket_model.WsClient) {
	for s := range c.GetSubscriptions() {
		err := cm.removeSubscription(s, c, true)

		// Client is leaving anyway, no point returning an error from here
		// just print it out
		if err != nil {
			cm.log.Error().Stack().Err(err).Msg("")
		}
	}

	cm.clientMu.Lock()
	defer cm.clientMu.Unlock()
	if c.GetUser() != nil {
		delete(cm.webClientMap, c.GetClientId())
	} else if remote := c.GetInstance(); remote != nil {
		delete(cm.remoteClientMap, c.GetInstance().TowerId)
		cm.PushWeblensEvent(websocket_mod.RemoteConnectionChangedEvent, websocket_mod.WsData{"remoteId": remote.TowerId, "online": false})
	}

	c.Disconnect()
}

func (cm *ClientManager) GetClientByUsername(username string) *websocket_model.WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()

	// FIXME: This is O(n) and should be better
	cm.log.Debug().Msgf("Checking at most %d clients", len(cm.webClientMap))
	for _, c := range cm.webClientMap {
		cm.log.Debug().Msgf("Checking client [%s] against [%s]", c.GetUser().GetUsername(), username)
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

func (cm *ClientManager) GetSubscribers(st websocket_mod.WsAction, key string) (clients []*websocket_model.WsClient) {
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
		cm.log.Error().Msgf("Unknown subscriber type: [%s]", st)
	}

	// Copy clients to not modify reference in the map
	return clients[:]
}

func (cm *ClientManager) SubscribeToFile(ctx context.NotifierContext, c *client.WsClient, file *file_model.WeblensFileImpl, share *share_model.FileShare, subTime time.Time) error {

	sub := websocket_mod.Subscription{Type: websocket_mod.FolderSubscribe, SubscriptionId: file.ID(), When: subTime}

	// TODO: Check if subscription is needed to be added to client
	// c.AddSubscription(sub)

	cm.addSubscription(sub, c)

	notifs := client.NewFileNotification(file, websocket_mod.FileUpdatedEvent, nil)
	ctx.Notify(notifs...)

	// for _, t := range cm.pack.GetFileService().GetTasks(folder) {
	// 	// c.SubUnlock()
	// 	_, _, err = cm.Subscribe(c, t.TaskId(), websocket_model.TaskSubscribe, time.Now(), nil)
	// 	// c.SubLock()
	// 	if err != nil {
	// 		return
	// 	}
	// }

	return nil
}

func (cm *ClientManager) SubscribeToTask(ctx context.NotifierContext, c *client.WsClient, task *task_model.Task, subTime time.Time) error {
	if done, _ := task.Status(); done {
		return nil
	}

	sub := websocket_mod.Subscription{Type: websocket_mod.TaskSubscribe, SubscriptionId: task.TaskId(), When: subTime}

	cm.addSubscription(sub, c)

	notif := client.NewTaskNotification(task, websocket_mod.TaskCreatedEvent, task.GetMeta().FormatToResult())
	ctx.Notify(notif)

	return nil
}

// Subscribe links a websocket connection to a "key" that can be broadcast to later if
// relevant updates need to be communicated
//
// Returns "true" and the results at meta.LookingFor if the task is completed, false and nil otherwise.
// Subscriptions to types that represent ongoing events like FolderSubscribe never return a truthy completed
//
// Deprecated: Use specific subscription types instead of this generic one.
func (cm *ClientManager) Subscribe(
	c *websocket_model.WsClient, key string, action websocket_mod.WsAction, subTime time.Time, share *share_model.FileShare,
) (complete bool, results map[*task_model.TaskResult]any, err error) {
	var sub websocket_mod.Subscription

	if c == nil {
		return false, nil, errors.New("Trying to subscribe nil client")
	}

	switch action {
	// case websocket_model.TaskSubscribe:
	// 	{
	// 		t := cm.pack.TaskService.GetTask(key)
	// 		if t == nil {
	// 			err = werror.Errorf("could not find task with ID %s", key)
	// 			c.Error(err)
	// 			return
	// 		}
	//
	// 		complete, _ = t.Status()
	// 		results = t.GetResults()
	//
	// 		if complete {
	// 			return
	// 		}
	//
	// 		for clientSub := range c.GetSubscriptions() {
	// 			if clientSub.SubscriptionId == key {
	// 				return
	// 			}
	// 		}
	//
	// 		sub = websocket_mod.Subscription{Type: websocket_model.TaskSubscribe, SubscriptionId: key, When: subTime}
	//
	// 		c.PushTaskUpdate(t, websocket_model.TaskCreatedEvent, t.GetMeta().FormatToResult())
	// 	}

	// TODO: Move this to its own function
	// case websocket_model.TaskTypeSubscribe:
	// 	{
	// 		sub = websocket_mod.Subscription{Type: websocket_model.TaskTypeSubscribe, SubscriptionId: key, When: subTime}
	// 	}
	default:
		{
			err = fmt.Errorf("unknown subscription type %s", action)
			cm.log.Error().Stack().Err(err).Msg("")
			c.Error(err)
			return
		}
	}

	c.AddSubscription(sub)
	cm.addSubscription(sub, c)

	return
}

func (cm *ClientManager) Unsubscribe(c *websocket_model.WsClient, key string, unSubTime time.Time) error {
	c.SubLock()
	defer c.SubUnlock()

	var sub websocket_mod.Subscription
	for s := range c.GetSubscriptions() {
		if s.SubscriptionId == key && !s.When.Before(unSubTime) {
			cm.log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Ignoring unsubscribe request that happened before subscribe request") })
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

func (cm *ClientManager) FolderSubToTask(folderId file_model.FileId, taskId task_model.Id) {
	subs := cm.GetSubscribers(websocket_mod.FolderSubscribe, folderId)

	for _, s := range subs {
		_, _, err := cm.Subscribe(s, taskId, websocket_mod.TaskSubscribe, time.Now(), nil)
		if err != nil {
			cm.log.Error().Stack().Err(err).Msg("")
		}
	}
}

func (cm *ClientManager) UnsubTask(taskId task_model.Id) {
	subs := cm.GetSubscribers(websocket_mod.TaskSubscribe, taskId)

	for _, s := range subs {
		err := cm.Unsubscribe(s, taskId, time.Now())
		if err != nil && !errors.Is(err, ErrSubscriptionNotFound) {
			cm.log.Error().Stack().Err(err).Msg("")
		} else if err != nil {
			cm.log.Warn().Msgf("Subscription [%s] not found in unsub task", taskId)
		}
	}
}

func (cm *ClientManager) Send(msg websocket_mod.WsResponseInfo) {
	defer internal.RecoverPanic("Panic caught while broadcasting")

	if msg.SubscribeKey == "" {
		cm.log.Error().Stack().Msg("Trying to broadcast on empty key")
		return
	}

	var clients []*websocket_model.WsClient

	if msg.BroadcastType == "serverEvent" {
		clients = cm.GetAllClients()
	} else {
		clients = cm.GetSubscribers(msg.BroadcastType, msg.SubscribeKey)
		clients = slices_mod.OnlyUnique(clients)

		if msg.BroadcastType == websocket_mod.TaskSubscribe {
			clients = append(
				clients, cm.GetSubscribers(
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
		cm.log.Trace().Str("websocket_event", string(msg.EventTag)).Msgf("Sending [%s] websocket message to %d client(s)", msg.EventTag, len(clients))
		for _, c := range clients {
			err := c.Send(msg)
			if err != nil {
				cm.log.Error().Stack().Err(err).Msg("")
			}
		}
	} else {
		cm.log.Trace().Func(func(e *zerolog.Event) {
			e.Msgf("No subscribers to [%s]. Trying to send [%s]", msg.SubscribeKey, msg.EventTag)
		})
		return
	}
}

func (cm *ClientManager) Relay(msg websocket_mod.WsResponseInfo) {

}

func (cm *ClientManager) PushWeblensEvent(event websocket_mod.WsEvent, msg websocket_mod.WsData) {
	cm.Send(websocket_mod.WsResponseInfo{
		EventTag:      event,
		Content:       msg,
		BroadcastType: "serverEvent",
	})
}

func (cm *ClientManager) addSubscription(subInfo websocket_mod.Subscription, client *websocket_model.WsClient) {
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
			cm.log.Error().Msgf("Unknown subType: %s", subInfo.Type)
			return
		}
	}
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
