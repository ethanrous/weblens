package service

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/task"
	"github.com/gorilla/websocket"
)

var _ models.ClientManager = (*ClientManager)(nil)

type ClientManager struct {
	webClientMap    map[models.Username]*models.WsClient
	remoteClientMap map[models.InstanceId]*models.WsClient

	core *models.WsClient

	// Key: subscription identifier, value: clientConn instance
	// Use map to take advantage of O(1) lookup time when finding or removing clients
	// by subscription identifier
	// {
	// 	"fileId": [
	// 		*client1,
	// 		*client2,
	//     ]
	// }
	folderSubs map[models.SubId][]*models.WsClient

	taskSubs map[models.SubId][]*models.WsClient

	taskTypeSubs map[models.SubId][]*models.WsClient

	pack     *models.ServicePack
	clientMu sync.RWMutex

	folderMu sync.Mutex

	taskMu sync.Mutex

	taskTypeMu sync.Mutex
}

func NewClientManager(
	pack *models.ServicePack,
) *ClientManager {
	cm := &ClientManager{
		webClientMap:    map[models.Username]*models.WsClient{},
		remoteClientMap: map[models.InstanceId]*models.WsClient{},

		folderSubs:   map[models.SubId][]*models.WsClient{},
		taskSubs:     map[models.SubId][]*models.WsClient{},
		taskTypeSubs: map[models.SubId][]*models.WsClient{},

		pack: pack,
	}

	return cm
}

func (cm *ClientManager) ClientConnect(conn *websocket.Conn, user *models.User) *models.WsClient {
	newClient := models.NewClient(conn, user)

	cm.clientMu.Lock()
	cm.webClientMap[newClient.GetClientId()] = newClient
	cm.clientMu.Unlock()

	log.Trace.Func(func(l log.Logger) { l.Printf("Web client [%s] connected", user.GetUsername()) })
	return newClient
}

func (cm *ClientManager) RemoteConnect(conn *websocket.Conn, remote *models.Instance) *models.WsClient {
	newClient := models.NewClient(conn, remote)

	cm.clientMu.Lock()
	cm.remoteClientMap[remote.ServerId()] = newClient
	cm.clientMu.Unlock()

	if remote.IsCore() {
		cm.core = newClient
	}

	log.Trace.Func(func(l log.Logger) { l.Printf("Server [%s] connected", remote.Name) })
	cm.pack.Caster.PushWeblensEvent(models.RemoteConnectionChangedEvent, models.WsC{"remoteId": remote.ServerId(), "online": true})
	return newClient
}

func (cm *ClientManager) ClientDisconnect(c *models.WsClient) {
	for s := range c.GetSubscriptions() {
		err := cm.removeSubscription(s, c, true)

		// Client is leaving anyway, no point returning an error from here
		// just print it out
		if err != nil {
			log.ErrTrace(err)
		}
	}

	cm.clientMu.Lock()
	defer cm.clientMu.Unlock()
	if c.GetUser() != nil {
		delete(cm.webClientMap, c.GetClientId())
	} else if remote := c.GetRemote(); remote != nil {
		delete(cm.remoteClientMap, c.GetRemote().ServerId())
		cm.pack.Caster.PushWeblensEvent(models.RemoteConnectionChangedEvent, models.WsC{"remoteId": remote.ServerId(), "online": false})
	}

	c.Disconnect()
}

func (cm *ClientManager) GetClientByUsername(username models.Username) *models.WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()

	for _, c := range cm.webClientMap {
		if c.GetUser().GetUsername() == username {
			return c
		}
	}

	return nil
}

func (cm *ClientManager) GetClientByServerId(instanceId models.InstanceId) *models.WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return cm.remoteClientMap[instanceId]
}

func (cm *ClientManager) GetAllClients() []*models.WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return append(internal.MapToValues(cm.webClientMap), internal.MapToValues(cm.remoteClientMap)...)
}

func (cm *ClientManager) GetConnectedAdmins() []*models.WsClient {
	clients := cm.GetAllClients()
	admins := internal.Filter(
		clients, func(c *models.WsClient) bool {
			return c.GetUser().IsAdmin()
		},
	)
	return admins
}

func (cm *ClientManager) GetSubscribers(st models.WsAction, key models.SubId) (clients []*models.WsClient) {
	switch st {
	case models.FolderSubscribe:
		{
			cm.folderMu.Lock()
			clients = cm.folderSubs[key]
			cm.folderMu.Unlock()
		}
	case models.TaskSubscribe:
		{
			cm.taskMu.Lock()
			clients = cm.taskSubs[key]
			cm.taskMu.Unlock()
		}
	case models.UserSubscribe:
		{
			cm.clientMu.Lock()
			allClients := internal.MapToValues(cm.webClientMap)
			cm.clientMu.Unlock()
			clients = internal.Filter(
				allClients, func(c *models.WsClient) bool {
					return c.GetUser().GetUsername() == key
				},
			)
		}
	case models.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			clients = cm.taskTypeSubs[key]
			cm.taskTypeMu.Unlock()
		}
	default:
		log.Error.Printf("Unknown subscriber type: [%s]", st)
	}

	// Copy clients to not modify reference in the map
	return clients[:]
}

// Subscribe links a websocket connection to a "key" that can be broadcast to later if
// relevant updates need to be communicated
//
// Returns "true" and the results at meta.LookingFor if the task is completed, false and nil otherwise.
// Subscriptions to types that represent ongoing events like FolderSubscribe never return a truthy completed
func (cm *ClientManager) Subscribe(
	c *models.WsClient, key models.SubId, action models.WsAction, subTime time.Time, share models.Share,
) (complete bool, results map[task.TaskResultKey]any, err error) {
	var sub models.Subscription

	if c == nil {
		return false, nil, werror.Errorf("Trying to subscribe nil client")
	}

	switch action {
	case models.FolderSubscribe:
		{
			if key == "" {
				err = fmt.Errorf("cannot subscribe with empty folder id")
				c.Error(err)
				return
			}
			fileId := key
			var folder *fileTree.WeblensFileImpl
			if fileId == "external" {
				fileId = "EXTERNAL"
			}

			var fileShare *models.FileShare
			if fsh, ok := share.(*models.FileShare); ok {
				fileShare = fsh
			}

			folder, err = cm.pack.GetFileService().GetFileSafe(fileId, c.GetUser(), fileShare)
			if err != nil {
				c.Error(err)
				return
			}

			sub = models.Subscription{Type: models.FolderSubscribe, Key: key, When: subTime}
			c.PushFileUpdate(folder, nil)

			for _, t := range cm.pack.GetFileService().GetTasks(folder) {
				// c.SubUnlock()
				_, _, err = cm.Subscribe(c, t.TaskId(), models.TaskSubscribe, time.Now(), nil)
				// c.SubLock()
				if err != nil {
					return
				}
			}
		}
	case models.TaskSubscribe:
		{
			t := cm.pack.TaskService.GetTask(key)
			if t == nil {
				err = werror.Errorf("could not find task with ID %s", key)
				c.Error(err)
				return
			}

			complete, _ = t.Status()
			results = t.GetResults()

			if complete {
				return
			}

			for clientSub := range c.GetSubscriptions() {
				if clientSub.Key == key {
					return
				}
			}

			sub = models.Subscription{Type: models.TaskSubscribe, Key: key, When: subTime}

			c.PushTaskUpdate(t, models.TaskCreatedEvent, t.GetMeta().FormatToResult())
		}
	case models.TaskTypeSubscribe:
		{
			sub = models.Subscription{Type: models.TaskTypeSubscribe, Key: key, When: subTime}
		}
	default:
		{
			err = fmt.Errorf("unknown subscription type %s", action)
			log.ErrTrace(err)
			c.Error(err)
			return
		}
	}

	log.Trace.Func(func(l log.Logger) { l.Printf("U[%s] subscribed to [%s]", c.GetUser().GetUsername(), key) })

	c.AddSubscription(sub)
	cm.addSubscription(sub, c)

	return
}

func (cm *ClientManager) Unsubscribe(c *models.WsClient, key models.SubId, unSubTime time.Time) error {
	c.SubLock()
	defer c.SubUnlock()

	var sub models.Subscription
	for s := range c.GetSubscriptions() {
		if s.Key == key && !s.When.Before(unSubTime) {
			log.Debug.Println("Ignoring unsubscribe request that happened before subscribe request")
			continue
		}

		if s.Key == key && s.When.Before(unSubTime) {
			sub = s
			break
		}
	}

	if sub == (models.Subscription{}) {
		return werror.WithStack(werror.ErrSubscriptionNotFound)
	}
	log.Trace.Func(func(l log.Logger) { l.Printf("Removing [%s]'s subscription to [%s]", c.GetUser().GetUsername(), key) })

	c.RemoveSubscription(key)
	return cm.removeSubscription(sub, c, false)
}

func (cm *ClientManager) FolderSubToTask(folderId fileTree.FileId, taskId task.Id) {
	subs := cm.GetSubscribers(models.FolderSubscribe, folderId)

	for _, s := range subs {
		log.Trace.Func(func(l log.Logger) {
			l.Printf(
				"Subscribing U[%s] to T[%s] due to F[%s]", s.GetUser().GetUsername(),
				taskId, folderId,
			)
		})
		_, _, err := cm.Subscribe(s, taskId, models.TaskSubscribe, time.Now(), nil)
		if err != nil {
			log.ShowErr(err)
		}
	}
}

func (cm *ClientManager) UnsubTask(taskId task.Id) {
	subs := cm.GetSubscribers(models.TaskSubscribe, taskId)

	for _, s := range subs {
		log.Debug.Func(func(l log.Logger) {
			l.Printf(
				"Unsubscribing U[%s] from T[%s]", s.GetUser().GetUsername(), taskId)
		})
		err := cm.Unsubscribe(s, taskId, time.Now())
		if err != nil && !errors.Is(err, werror.ErrSubscriptionNotFound) {
			log.ShowErr(err)
		} else if err != nil {
			log.Warning.Printf("Subscription [%s] not found in unsub task", taskId)
		}
	}
}

func (cm *ClientManager) Send(msg models.WsResponseInfo) {
	defer internal.RecoverPanic("Panic caught while broadcasting")

	if msg.SubscribeKey == "" {
		log.Error.Println("Trying to broadcast on empty key")
		return
	}

	var clients []*models.WsClient

	if msg.BroadcastType == "serverEvent" || cm.pack.GetFileService() == nil || cm.pack.InstanceService.GetLocal().GetRole() == models.BackupServerRole {
		clients = cm.GetAllClients()
	} else {
		clients = cm.GetSubscribers(msg.BroadcastType, msg.SubscribeKey)
		clients = internal.OnlyUnique(clients)

		if msg.BroadcastType == models.TaskSubscribe {
			clients = append(
				clients, cm.GetSubscribers(
					models.TaskTypeSubscribe,
					msg.TaskType,
				)...,
			)
		}
	}

	if len(clients) != 0 {
		for _, c := range clients {
			err := c.Send(msg)
			if err != nil {
				log.ErrTrace(err)
			}
		}
	} else {
		// log.TraceCaller(2, "No subscribers to [%s]", msg.SubscribeKey)
		log.Trace.Func(func(l log.Logger) {
			l.Printf("No subscribers to [%s]. Trying to send [%s]", msg.SubscribeKey, msg.EventTag)
		})
		return
	}
}

func (cm *ClientManager) addSubscription(subInfo models.Subscription, client *models.WsClient) {
	switch subInfo.Type {
	case models.FolderSubscribe:
		{
			cm.folderMu.Lock()
			addSub(cm.folderSubs, subInfo, client)
			cm.folderMu.Unlock()
		}
	case models.TaskSubscribe:
		{
			cm.taskMu.Lock()
			addSub(cm.taskSubs, subInfo, client)
			cm.taskMu.Unlock()
		}
	case models.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			addSub(cm.taskTypeSubs, subInfo, client)
			cm.taskTypeMu.Unlock()
		}
	default:
		{
			log.Error.Println("Unknown subType", subInfo.Type)
			return
		}
	}
}

func (cm *ClientManager) removeSubscription(
	subInfo models.Subscription, client *models.WsClient, removeAll bool,
) error {
	var err error
	switch subInfo.Type {
	case models.FolderSubscribe:
		{
			cm.folderMu.Lock()
			err = removeSubs(cm.folderSubs, subInfo, client, removeAll)
			cm.folderMu.Unlock()
		}
	case models.TaskSubscribe:
		{
			cm.taskMu.Lock()
			err = removeSubs(cm.taskSubs, subInfo, client, removeAll)
			cm.taskMu.Unlock()
		}
	case models.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			err = removeSubs(cm.taskTypeSubs, subInfo, client, removeAll)
			cm.taskTypeMu.Unlock()
		}
	default:
		{
			return werror.Errorf("Trying to remove unknown subscription type [%s]", subInfo.Type)
		}
	}

	return err
}

func addSub(subMap map[models.SubId][]*models.WsClient, subInfo models.Subscription, client *models.WsClient) {
	subs, ok := subMap[subInfo.Key]

	if !ok {
		subs = []*models.WsClient{}
	}

	subMap[subInfo.Key] = append(subs, client)
}

func removeSubs(
	subMap map[models.SubId][]*models.WsClient, subInfo models.Subscription, client *models.WsClient, removeAll bool,
) error {
	subs, ok := subMap[subInfo.Key]
	if !ok {
		return werror.Errorf("Tried to unsubscribe from non-existent key [%s]", subInfo.Key)
	}
	if removeAll {
		subs = internal.Filter(subs, func(c *models.WsClient) bool { return c.GetClientId() != client.GetClientId() })
	} else {
		index := slices.IndexFunc(
			subs, func(c *models.WsClient) bool { return c.GetClientId() == client.GetClientId() },
		)
		if index != -1 {
			subs = internal.Banish(subs, index)
		}
	}
	subMap[subInfo.Key] = subs

	return nil
}
