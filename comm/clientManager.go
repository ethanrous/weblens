package comm

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/models/service"
	"github.com/ethrousseau/weblens/task"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var _ ClientManager = (*clientManager)(nil)

type clientManager struct {
	webClientMap    map[models.Username]*WsClient
	remoteClientMap map[models.InstanceId]*WsClient
	clientMu        *sync.RWMutex

	core *WsClient

	// Key: subscription identifier, value: clientConn instance
	// Use map to take advantage of O(1) lookup time when finding or removing clients
	// by subscription identifier
	// {
	// 	"fileId": [
	// 		*client1,
	// 		*client2,
	//     ]
	// }
	folderSubs   map[SubId][]*WsClient
	taskSubs     map[SubId][]*WsClient
	taskTypeSubs map[SubId][]*WsClient
	folderMu     sync.Mutex
	taskMu       sync.Mutex
	taskTypeMu   sync.Mutex

	fileService *service.FileServiceImpl
	taskService task.TaskService
}

func NewClientManager(fileService *service.FileServiceImpl, taskService task.TaskService) ClientManager {
	return &clientManager{
		webClientMap:    map[models.Username]*WsClient{},
		remoteClientMap: map[models.InstanceId]*WsClient{},
		clientMu:        &sync.RWMutex{},

		folderSubs:   map[SubId][]*WsClient{},
		taskSubs:     map[SubId][]*WsClient{},
		taskTypeSubs: map[SubId][]*WsClient{},

		fileService: fileService,
		taskService: taskService,
	}
}

func (cm *clientManager) ClientConnect(conn *websocket.Conn, user *models.User) *WsClient {
	connectionId := ClientId(uuid.New().String())
	newClient := WsClient{Active: true, connId: connectionId, conn: conn, user: user}

	cm.clientMu.Lock()
	cm.webClientMap[user.GetUsername()] = &newClient
	cm.clientMu.Unlock()

	log.Debug.Printf("Websocket client [%s] connected", user.GetUsername())
	return &newClient
}

func (cm *clientManager) RemoteConnect(conn *websocket.Conn, remote *models.Instance) *WsClient {
	connectionId := ClientId(uuid.New().String())
	newClient := &WsClient{Active: true, conn: conn, remote: remote, connId: connectionId}

	cm.clientMu.Lock()
	cm.remoteClientMap[remote.ServerId()] = newClient
	cm.clientMu.Unlock()

	if remote.IsCore() {
		cm.core = newClient
	}

	return newClient
}

func (cm *clientManager) ClientDisconnect(c *WsClient) {
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

	c.disconnect()
}

func (cm *clientManager) GetClientByUsername(username models.Username) *WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return cm.webClientMap[username]
}

func (cm *clientManager) GetClientByInstanceId(instanceId models.InstanceId) *WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return cm.remoteClientMap[instanceId]
}

func (cm *clientManager) GetAllClients() []*WsClient {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return internal.MapToValues(cm.webClientMap)
}

func (cm *clientManager) GetConnectedAdmins() []*WsClient {
	clients := cm.GetAllClients()
	admins := internal.Filter(
		clients, func(c *WsClient) bool {
			return c.GetUser().IsAdmin()
		},
	)
	return admins
}

func (cm *clientManager) GetSubscribers(st WsAction, key SubId) (clients []*WsClient) {
	switch st {
	case FolderSubscribe:
		{
			cm.folderMu.Lock()
			clients = cm.folderSubs[key]
			cm.folderMu.Unlock()
		}
	case TaskSubscribe:
		{
			cm.taskMu.Lock()
			clients = cm.taskSubs[key]
			cm.taskMu.Unlock()
		}
	case UserSubscribe:
		{
			cm.clientMu.Lock()
			allClients := internal.MapToValues(cm.webClientMap)
			cm.clientMu.Unlock()
			clients = internal.Filter(
				allClients, func(c *WsClient) bool {
					return SubId(c.GetUser().GetUsername()) == key
				},
			)
		}
	case TaskTypeSubscribe:
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
func (cm *clientManager) Subscribe(
	c *WsClient,
	key SubId, action WsAction, share models.Share,
) (
	complete bool,
	results map[string]any,
	err error,
) {
	var sub Subscription

	// *HACK* Ensure that subscribe requests are processed after unsubscribe requests
	// that are sent at the same time from the client. A lower sleep value may be able
	// to achieve the same effect, but this works for now...
	time.Sleep(time.Millisecond * 100)

	c.SubLock()
	defer c.SubUnlock()

	switch action {
	case FolderSubscribe:
		{
			if key == "" {
				err = fmt.Errorf("cannot subscribe with empty folder id")
				c.Error(err)
				return
			}
			fileId := fileTree.FileId(key)
			var folder *fileTree.WeblensFile
			if fileId == "external" {
				fileId = "EXTERNAL"
			}

			var fileShare *models.FileShare
			if fsh, ok := share.(*models.FileShare); ok {
				fileShare = fsh
			}

			folder, err = cm.fileService.GetFileSafe(fileId, c.GetUser(), fileShare)
			if err != nil {
				safe, _ := werror.TrySafeErr(err)
				c.Error(safe)
				// c.Error(werror.Errorf("failed to find folder with id [%s] to subscribe to", fileId))
				return
			}

			sub = Subscription{Type: FolderSubscribe, Key: key}
			c.PushFileUpdate(folder, nil)

			for _, t := range cm.fileService.GetTasks(folder) {
				c.SubUnlock()
				_, _, err = cm.Subscribe(c, SubId(t.TaskId()), TaskSubscribe, nil)
				c.SubLock()
				if err != nil {
					return
				}
			}
			// TODO
		}
	case TaskSubscribe:
		{
			t := cm.taskService.GetTask(task.TaskId(key))
			if t == nil {
				err = fmt.Errorf("could not find task with ID %s", key)
				c.Error(err)
				return
			}

			complete, _ = t.Status()
			results = t.GetResults()

			c.updateMu.Lock()
			if complete || slices.IndexFunc(
				c.subscriptions, func(s Subscription) bool { return s.Key == key },
			) != -1 {
				c.updateMu.Unlock()
				return
			}
			c.updateMu.Unlock()

			sub = Subscription{Type: TaskSubscribe, Key: key}

			c.PushTaskUpdate(t, TaskCreatedEvent, t.GetMeta().FormatToResult())
		}
	case PoolSubscribe:
		{
			pool := cm.taskService.GetTaskPool(task.TaskId(key))
			if pool == nil {
				c.Error(errors.New(fmt.Sprintf("Could not find pool with id %s", key)))
				return
			} else if pool.IsGlobal() {
				c.Error(errors.New("Trying to subscribe to global pool"))
				return
			}

			log.Debug.Printf("%s subscribed to pool [%s]", c.GetUser().GetUsername(), pool.ID())
			sub = Subscription{Type: TaskSubscribe, Key: key}

			c.PushPoolUpdate(
				pool, PoolCreatedEvent, task.TaskResult{
					"createdBy": pool.CreatedInTask().
						TaskId(),
				},
			)
		}
	case TaskTypeSubscribe:
		{
			log.Debug.Printf("%s subscribed to task type [%s]", c.user.GetUsername(), key)
			sub = Subscription{Type: TaskTypeSubscribe, Key: key}
		}
	default:
		{
			err = fmt.Errorf("unknown subscription type %s", action)
			log.ErrTrace(err)
			c.Error(err)
			return
		}
	}

	log.Debug.Printf("[%s] subscribed to %s", c.user.GetUsername(), key)

	c.updateMu.Lock()
	c.subscriptions = append(c.subscriptions, sub)
	c.updateMu.Unlock()
	cm.addSubscription(sub, c)

	return
}

func (cm *clientManager) Unsubscribe(c *WsClient, key SubId) error {
	c.SubLock()
	defer c.SubUnlock()

	var sub Subscription
	for s := range c.GetSubscriptions() {
		if s.Key == key {
			sub = s
			break
		}
	}

	if sub == (Subscription{}) {
		return werror.Errorf("Could not find subscription with key [%s]", key)
	}
	log.Debug.Printf("Removing [%s]'s subscription to [%s]", c.user.GetUsername(), key)

	return cm.removeSubscription(sub, c, false)
}

func (cm *clientManager) addSubscription(subInfo Subscription, client *WsClient) {
	switch subInfo.Type {
	case FolderSubscribe:
		{
			cm.folderMu.Lock()
			addSub(cm.folderSubs, subInfo, client)
			cm.folderMu.Unlock()
		}
	case TaskSubscribe:
		{
			cm.taskMu.Lock()
			addSub(cm.taskSubs, subInfo, client)
			cm.taskMu.Unlock()
		}
	case TaskTypeSubscribe:
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

func (cm *clientManager) FolderSubToPool(folderId fileTree.FileId, poolId task.TaskId) {
	subs := cm.GetSubscribers(FolderSubscribe, SubId(folderId))

	for _, s := range subs {
		log.Debug.Printf("Subscribing user %s on folder sub %s to pool %s", s.user.GetUsername(), folderId, poolId)
		_, _, err := cm.Subscribe(s, SubId(poolId), PoolSubscribe, nil)
		if err != nil {
			log.ShowErr(err)
		}
	}
}

func (cm *clientManager) removeSubscription(subInfo Subscription, client *WsClient, removeAll bool) error {
	var err error
	switch subInfo.Type {
	case FolderSubscribe:
		{
			cm.folderMu.Lock()
			err = removeSubs(cm.folderSubs, subInfo, client, removeAll)
			cm.folderMu.Unlock()
		}
	case TaskSubscribe:
		{
			cm.taskMu.Lock()
			err = removeSubs(cm.taskSubs, subInfo, client, removeAll)
			cm.taskMu.Unlock()
		}
	case TaskTypeSubscribe:
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

func (cm *clientManager) Send(msg WsResponseInfo) {
	defer internal.RecoverPanic("Panic caught while broadcasting")

	if msg.SubscribeKey == "" {
		log.Error.Println("Trying to broadcast on empty key")
		return
	}

	var clients []*WsClient
	if !InstanceService.IsLocalLoaded() || msg.BroadcastType == ServerEvent {
		clients = cm.GetAllClients()
	} else {
		clients = cm.GetSubscribers(msg.BroadcastType, msg.SubscribeKey)
		clients = internal.OnlyUnique(clients)
	}

	if msg.BroadcastType == TaskSubscribe {
		clients = append(
			clients, cm.GetSubscribers(
				TaskTypeSubscribe,
				SubId(msg.TaskType),
			)...,
		)
	}

	if len(clients) != 0 {
		for _, c := range clients {
			c.send(msg)
		}
	} else {
		// Although debug is our "verbose" mode, this one is *really* annoying, so it's disabled unless needed.
		// util.Debug.Println("No subscribers to", msg.SubscribeKey)
		return
	}
}

func addSub(subMap map[SubId][]*WsClient, subInfo Subscription, client *WsClient) {
	subs, ok := subMap[subInfo.Key]

	if !ok {
		subs = []*WsClient{}
	}

	subMap[subInfo.Key] = append(subs, client)
}

func removeSubs(
	subMap map[SubId][]*WsClient, subInfo Subscription, client *WsClient, removeAll bool,
) error {
	subs, ok := subMap[subInfo.Key]
	if !ok {
		return werror.Errorf("Tried to unsubscribe from non-existent key [%s]", subInfo.Key)
	}
	if removeAll {
		subs = internal.Filter(subs, func(c *WsClient) bool { return c.GetClientId() != client.GetClientId() })
	} else {
		index := slices.IndexFunc(subs, func(c *WsClient) bool { return c.GetClientId() == client.GetClientId() })
		if index != -1 {
			subs = internal.Banish(subs, index)
		}
	}
	subMap[subInfo.Key] = subs

	return nil
}

type WsAuthorize struct {
	Auth string `json:"auth"`
}

const RetryInterval = time.Second * 10

func WebsocketToCore(core *models.Instance, clientService *clientManager) error {
	addrStr, err := core.GetAddress()
	if err != nil {
		return err
	}

	if addrStr == "" {
		return errors.New("Core server address is empty")
	}

	re, err := regexp.Compile(`http(s)?://([^/]*)`)
	if err != nil {
		return werror.WithStack(err)
	}

	parts := re.FindStringSubmatch(addrStr)

	addr := flag.String("addr", parts[2], "http service address")
	host := url.URL{Scheme: "ws" + parts[1], Host: *addr, Path: "/api/core/ws"}
	dialer := &websocket.Dialer{Proxy: http.ProxyFromEnvironment, HandshakeTimeout: 10 * time.Second}

	authHeader := http.Header{}
	authHeader.Add("Authorization", "Bearer "+string(core.GetUsingKey()))
	var conn *WsClient
	go func() {
		for {
			conn, err = dial(dialer, host, authHeader, core, clientService)
			if err != nil {
				log.Warning.Printf(
					"Failed to connect to core server at %s, trying again in %s",
					host.String(), RetryInterval,
				)
				log.Debug.Println("Error was", err)
				time.Sleep(RetryInterval)
				continue
			}
			coreWsHandler(conn)
			log.Warning.Printf("Connection to core websocket closed, reconnecting...")
		}
	}()
	return nil
}

func dial(
	dialer *websocket.Dialer, host url.URL, authHeader http.Header, core *models.Instance,
	clientService *clientManager,
) (
	*WsClient, error,
) {
	log.Debug.Println("Dialing", host.String())
	conn, _, err := dialer.Dial(host.String(), authHeader)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	client := clientService.RemoteConnect(conn, core)

	err = client.Raw(WsAuthorize{Auth: authHeader.Get("Authorization")})
	if err != nil {
		return nil, werror.WithStack(err)
	}

	log.Info.Printf("Connection to core server at %s successfully established", host.String())
	return client, nil
}

func coreWsHandler(c *WsClient) {
	defer func() { c.disconnect() }()
	defer func() { recover() }()

	for {
		mt, message, err := c.ReadOne()
		if err != nil {
			log.ShowErr(werror.WithStack(err))
			break
		}
		log.Debug.Println(mt, string(message))
	}
}
