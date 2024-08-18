package routes

import (
	"slices"
	"sync"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/ethrousseau/weblens/api/util/wlog"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type clientManager struct {
	webClientMap    map[types.Username]types.Client
	remoteClientMap map[types.InstanceId]types.Client
	clientMu        *sync.RWMutex

	core types.Client

	// Key: subscription identifier, value: clientConn instance
	// Use map to take advantage of O(1) lookup time when finding or removing clients
	// by subscription identifier
	// {
	// 	"fileId": [
	// 		*client1,
	// 		*client2,
	//     ]
	// }
	folderSubs   map[types.SubId][]types.Client
	taskSubs     map[types.SubId][]types.Client
	taskTypeSubs map[types.SubId][]types.Client
	folderMu     sync.Mutex
	taskMu       sync.Mutex
	taskTypeMu   sync.Mutex
}

func NewClientManager() types.ClientManager {
	return &clientManager{
		webClientMap:    map[types.Username]types.Client{},
		remoteClientMap: map[types.InstanceId]types.Client{},
		clientMu:        &sync.RWMutex{},

		folderSubs:   map[types.SubId][]types.Client{},
		taskSubs:     map[types.SubId][]types.Client{},
		taskTypeSubs: map[types.SubId][]types.Client{},

		// folderMu: sync.Mutex{},
		// taskMu:   sync.Mutex{},
	}
}

func (cm *clientManager) ClientConnect(conn *websocket.Conn, user types.User) types.Client {
	connectionId := types.ClientId(uuid.New().String())
	newClient := clientConn{Active: true, connId: connectionId, conn: conn, user: user}

	cm.clientMu.Lock()
	cm.webClientMap[user.GetUsername()] = &newClient
	cm.clientMu.Unlock()

	newClient.debug("Connected")
	return &newClient
}

func (cm *clientManager) RemoteConnect(conn *websocket.Conn, remote types.Instance) types.Client {
	connectionId := types.ClientId(uuid.New().String())
	newClient := &clientConn{Active: true, conn: conn, remote: remote, connId: connectionId}

	cm.clientMu.Lock()
	cm.remoteClientMap[remote.ServerId()] = newClient
	cm.clientMu.Unlock()

	if remote.IsCore() {
		cm.core = newClient
	}

	return newClient
}

func (cm *clientManager) ClientDisconnect(c types.Client) {
	for _, s := range c.GetSubscriptions() {
		cm.RemoveSubscription(s, c, true)
	}

	cm.clientMu.Lock()
	defer cm.clientMu.Unlock()
	if c.GetUser() != nil {
		delete(cm.webClientMap, c.GetUser().GetUsername())
	} else if c.GetRemote() != nil {
		delete(cm.remoteClientMap, c.GetRemote().ServerId())
	}
}

func (cm *clientManager) GetClientByUsername(username types.Username) types.Client {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return cm.webClientMap[username]
}

func (cm *clientManager) GetClientByInstanceId(instanceId types.InstanceId) types.Client {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return cm.remoteClientMap[instanceId]
}

func (cm *clientManager) GetAllClients() []types.Client {
	cm.clientMu.RLock()
	defer cm.clientMu.RUnlock()
	return util.MapToValues(cm.webClientMap)
}

func (cm *clientManager) GetConnectedAdmins() []types.Client {
	clients := cm.GetAllClients()
	admins := util.Filter(
		clients, func(c types.Client) bool {
			return c.GetUser().IsAdmin()
		},
	)
	return admins
}

func (cm *clientManager) GetSubscribers(st types.WsAction, key types.SubId) (clients []types.Client) {
	switch st {
	case types.FolderSubscribe:
		{
			cm.folderMu.Lock()
			clients = cm.folderSubs[key]
			cm.folderMu.Unlock()
		}
	case types.TaskSubscribe:
		{
			cm.taskMu.Lock()
			clients = cm.taskSubs[key]
			cm.taskMu.Unlock()
		}
	case types.UserSubscribe:
		{
			cm.clientMu.Lock()
			allClients := util.MapToValues(cm.webClientMap)
			cm.clientMu.Unlock()
			clients = util.Filter(
				allClients, func(c types.Client) bool {
					return types.SubId(c.GetUser().GetUsername()) == key
				},
			)
		}
	case types.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			clients = cm.taskTypeSubs[key]
			cm.taskTypeMu.Unlock()
		}
	default:
		wlog.Error.Printf("Unknown subscriber type: [%s]", st)
	}

	// Copy slice to not modify reference to mapped slice
	return clients[:]
}

func (cm *clientManager) AddSubscription(subInfo types.Subscription, client types.Client) {
	switch subInfo.Type {
	case types.FolderSubscribe:
		{
			cm.folderMu.Lock()
			addSub(cm.folderSubs, subInfo, client)
			cm.folderMu.Unlock()
		}
	case types.TaskSubscribe:
		{
			cm.taskMu.Lock()
			addSub(cm.taskSubs, subInfo, client)
			cm.taskMu.Unlock()
		}
	case types.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			addSub(cm.taskTypeSubs, subInfo, client)
			cm.taskTypeMu.Unlock()
		}
	default:
		{
			wlog.Error.Println("Unknown subType", subInfo.Type)
			return
		}
	}
}

func (cm *clientManager) RemoveSubscription(subInfo types.Subscription, client types.Client, removeAll bool) {
	switch subInfo.Type {
	case types.FolderSubscribe:
		{
			cm.folderMu.Lock()
			removeSubs(cm.folderSubs, subInfo, client, removeAll)
			cm.folderMu.Unlock()
		}
	case types.TaskSubscribe:
		{
			cm.taskMu.Lock()
			removeSubs(cm.taskSubs, subInfo, client, removeAll)
			cm.taskMu.Unlock()
		}
	case types.TaskTypeSubscribe:
		{
			cm.taskTypeMu.Lock()
			removeSubs(cm.taskTypeSubs, subInfo, client, removeAll)
			cm.taskTypeMu.Unlock()
		}
	default:
		{
			wlog.Error.Println("Unknown subType", string(subInfo.Type))
			return
		}
	}
}

func addSub(subMap map[types.SubId][]types.Client, subInfo types.Subscription, client types.Client) {
	subs, ok := subMap[subInfo.Key]

	if !ok {
		subs = []types.Client{}
	}

	subMap[subInfo.Key] = append(subs, client)
}

func removeSubs(
	subMap map[types.SubId][]types.Client, subInfo types.Subscription, client types.Client, removeAll bool,
) {
	subs, ok := subMap[subInfo.Key]
	if !ok {
		wlog.Warning.Println("Tried to unsubscribe from non-existent key", string(subInfo.Key))
		return
	}
	if removeAll {
		subs = util.Filter(subs, func(c types.Client) bool { return c.GetClientId() != client.GetClientId() })
	} else {
		index := slices.IndexFunc(subs, func(c types.Client) bool { return c.GetClientId() == client.GetClientId() })
		if index != -1 {
			subs = util.Banish(subs, index)
		}
	}
	subMap[subInfo.Key] = subs
}
