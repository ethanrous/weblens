package service

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
	"github.com/gorilla/websocket"
)

var _ models.ClientManager = (*ClientManager)(nil)

type ClientManager struct {
	webClientMap    map[models.Username]*models.WsClient
	remoteClientMap map[models.InstanceId]*models.WsClient
	clientMu        *sync.RWMutex

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
	folderSubs   map[models.SubId][]*models.WsClient
	taskSubs     map[models.SubId][]*models.WsClient
	taskTypeSubs map[models.SubId][]*models.WsClient
	folderMu     sync.Mutex
	taskMu       sync.Mutex
	taskTypeMu   sync.Mutex

	fileService     *FileServiceImpl
	taskService     task.TaskService
	instanceService models.InstanceService
}

func NewClientManager(
	fileService *FileServiceImpl, taskService task.TaskService,
	instanceService models.InstanceService,
) *ClientManager {
	return &ClientManager{
		webClientMap:    map[models.Username]*models.WsClient{},
		remoteClientMap: map[models.InstanceId]*models.WsClient{},
		clientMu:        &sync.RWMutex{},

		folderSubs:   map[models.SubId][]*models.WsClient{},
		taskSubs:     map[models.SubId][]*models.WsClient{},
		taskTypeSubs: map[models.SubId][]*models.WsClient{},

		fileService:     fileService,
		taskService:     taskService,
		instanceService: instanceService,
	}
}

func (cm *ClientManager) SetFileService(fileService *FileServiceImpl) {
	cm.fileService = fileService
}

func (cm *ClientManager) ClientConnect(conn *websocket.Conn, user *models.User) *models.WsClient {
	newClient := models.NewClient(conn, user)

	cm.clientMu.Lock()
	cm.webClientMap[user.GetUsername()] = newClient
	cm.clientMu.Unlock()

	log.Trace.Printf("Web client [%s] connected", user.GetUsername())
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

	log.Trace.Printf("Server [%s] connected", remote.Name)
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
		delete(cm.webClientMap, c.GetUser().GetUsername())
	} else if c.GetRemote() != nil {
		delete(cm.remoteClientMap, c.GetRemote().ServerId())
	}

	c.Disconnect()
}

func (cm *ClientManager) GetClientByUsername(username models.Username) *models.WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return cm.webClientMap[username]
}

func (cm *ClientManager) GetClientByInstanceId(instanceId models.InstanceId) *models.WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return cm.remoteClientMap[instanceId]
}

func (cm *ClientManager) GetAllClients() []*models.WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return internal.MapToValues(cm.webClientMap)
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
					return models.SubId(c.GetUser().GetUsername()) == key
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

	// Copy slice to not modify reference to mapped slice
	return clients[:]
}

// Subscribe links a websocket connection to a "key" that can be broadcast to later if
// relevant updates need to be communicated
//
// Returns "true" and the results at meta.LookingFor if the task is completed, false and nil otherwise.
// Subscriptions to types that represent ongoing events like FolderSubscribe never return a truthy completed
func (cm *ClientManager) Subscribe(
	c *models.WsClient, key models.SubId, action models.WsAction, subTime time.Time, share models.Share,
) (complete bool, results map[string]any, err error,) {
	var sub models.Subscription

	// *HACK* Ensure that subscribe requests are processed after unsubscribe requests
	// that are sent at the same time from the client. A lower sleep value may be able
	// to achieve the same effect, but this works for now...
	// time.Sleep(time.Millisecond * 100)

	switch action {
	case models.FolderSubscribe:
		{
			if key == "" {
				err = fmt.Errorf("cannot subscribe with empty folder id")
				c.Error(err)
				return
			}
			fileId := fileTree.FileId(key)
			var folder *fileTree.WeblensFileImpl
			if fileId == "external" {
				fileId = "EXTERNAL"
			}

			var fileShare *models.FileShare
			if fsh, ok := share.(*models.FileShare); ok {
				fileShare = fsh
			}

			folder, err = cm.fileService.GetFileSafe(fileId, c.GetUser(), fileShare)
			if err != nil {
				c.Error(err)
				return
			}

			sub = models.Subscription{Type: models.FolderSubscribe, Key: key, When: subTime}
			c.PushFileUpdate(folder, nil)

			for _, t := range cm.fileService.GetTasks(folder) {
				c.SubUnlock()
				_, _, err = cm.Subscribe(c, models.SubId(t.TaskId()), models.TaskSubscribe, time.Now(), nil)
				c.SubLock()
				if err != nil {
					return
				}
			}
			// TODO
		}
	case models.TaskSubscribe:
		{
			t := cm.taskService.GetTask(task.TaskId(key))
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
	case models.PoolSubscribe:
		{
			pool := cm.taskService.GetTaskPool(task.TaskId(key))
			if pool == nil {
				c.Error(errors.New(fmt.Sprintf("Could not find pool with id %s", key)))
				return
			} else if pool.IsGlobal() {
				c.Error(errors.New("Trying to subscribe to global pool"))
				return
			}

			sub = models.Subscription{Type: models.TaskSubscribe, Key: key, When: subTime}

			c.PushPoolUpdate(
				pool, models.PoolCreatedEvent, task.TaskResult{
					"createdBy": pool.CreatedInTask().
						TaskId(),
				},
			)
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

	log.Trace.Printf("[%s] subscribed to [%s]", c.GetUser().GetUsername(), key)

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
		return werror.Errorf("Could not find subscription with key [%s]", key)
	}
	log.Trace.Printf("Removing [%s]'s subscription to [%s]", c.GetUser().GetUsername(), key)

	return cm.removeSubscription(sub, c, false)
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

func (cm *ClientManager) FolderSubToPool(folderId fileTree.FileId, poolId task.TaskId) {
	subs := cm.GetSubscribers(models.FolderSubscribe, models.SubId(folderId))

	for _, s := range subs {
		log.Trace.Printf("Subscribing user %s on folder sub %s to pool %s", s.GetUser().GetUsername(), folderId, poolId)
		_, _, err := cm.Subscribe(s, models.SubId(poolId), models.PoolSubscribe, time.Now(), nil)
		if err != nil {
			log.ShowErr(err)
		}
	}
}

func (cm *ClientManager) TaskSubToPool(taskId task.TaskId, poolId task.TaskId) {
	subs := cm.GetSubscribers(models.TaskSubscribe, models.SubId(taskId))

	for _, s := range subs {
		log.Trace.Printf("Subscribing user %s on folder sub %s to pool %s", s.GetUser().GetUsername(), taskId, poolId)
		_, _, err := cm.Subscribe(s, models.SubId(poolId), models.PoolSubscribe, time.Now(), nil)
		if err != nil {
			log.ShowErr(err)
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

func (cm *ClientManager) Send(msg models.WsResponseInfo) {
	defer internal.RecoverPanic("Panic caught while broadcasting")

	if msg.SubscribeKey == "" {
		log.Error.Println("Trying to broadcast on empty key")
		return
	}

	var clients []*models.WsClient

	if msg.BroadcastType == models.ServerEvent {
		clients = cm.GetAllClients()
	} else {
		clients = cm.GetSubscribers(msg.BroadcastType, msg.SubscribeKey)
		clients = internal.OnlyUnique(clients)
	}

	if msg.BroadcastType == models.TaskSubscribe {
		clients = append(
			clients, cm.GetSubscribers(
				models.TaskTypeSubscribe,
				models.SubId(msg.TaskType),
			)...,
		)
	}

	if len(clients) != 0 {
		for _, c := range clients {
			err := c.Send(msg)
			if err != nil {
				log.ErrTrace(err)
			}
		}
	} else {
		// Although debug is our "verbose" mode, this one is *really* annoying, so it's disabled unless needed.
		// util.Debug.Println("No subscribers to", msg.SubscribeKey)
		return
	}
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
