package routes

import (
	"slices"
	"sync"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type clientManager struct {
	// Key: connection id, value: client instance
	clientMap map[types.Username]types.Client
	clientMu  *sync.Mutex

	// Key: subscription identifier, value: client instance
	// Use map to take advantage of O(1) lookup time when finding or removing clients
	// by subscription identifier
	// {
	// 	"fileId": [
	// 		*client1,
	// 		*client2,
	//     ]
	// }
	folderSubs map[types.SubId][]types.Client
	taskSubs   map[types.SubId][]types.Client
	folderMu   *sync.Mutex
	taskMu     *sync.Mutex
}

func NewClientManager() types.ClientManager {
	return &clientManager{
		clientMap: map[types.Username]types.Client{},
		clientMu:  &sync.Mutex{},

		folderSubs: map[types.SubId][]types.Client{},
		taskSubs:   map[types.SubId][]types.Client{},

		folderMu: &sync.Mutex{},
		taskMu:   &sync.Mutex{},
	}
}

func (cm *clientManager) ClientConnect(conn *websocket.Conn, user types.User) types.Client {
	connectionId := types.ClientId(uuid.New().String())
	newClient := client{Active: true, connId: connectionId, conn: conn, user: user}

	cm.clientMu.Lock()
	cm.clientMap[user.GetUsername()] = &newClient
	cm.clientMu.Unlock()

	newClient.debug("Connected")
	return &newClient
}

func (cm *clientManager) ClientDisconnect(c types.Client) {
	for _, s := range c.GetSubscriptions() {
		cm.RemoveSubscription(s, c, true)
	}

	cm.clientMu.Lock()
	delete(cm.clientMap, c.GetUser().GetUsername())
	cm.clientMu.Unlock()
}

func (cm *clientManager) GetClientByUsername(username types.Username) types.Client {
	cm.clientMu.Lock()
	defer cm.clientMu.Unlock()
	return cm.clientMap[username]
}

func (cm *clientManager) GetAllClients() []types.Client {
	cm.clientMu.Lock()
	defer cm.clientMu.Unlock()
	return util.MapToValues(cm.clientMap)
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
			allClients := util.MapToValues(cm.clientMap)
			cm.clientMu.Unlock()
			clients = util.Filter(
				allClients, func(c types.Client) bool {
					return types.SubId(c.GetUser().GetUsername()) == key
				},
			)
		}
	default:
		util.Error.Println("Unknown subscriber type", st)
	}

	// Copy slice to not modify reference to mapped slice
	return clients[0:]
}

func (cm *clientManager) AddSubscription(subInfo types.Subscription, client types.Client) {
	var subMap *map[types.SubId][]types.Client
	var lock *sync.Mutex

	switch subInfo.Type {
	case types.FolderSubscribe:
		{
			subMap = &cm.folderSubs
			lock = cm.folderMu
		}
	case types.TaskSubscribe:
		{
			subMap = &cm.taskSubs
			lock = cm.taskMu
		}
	default:
		{
			util.Error.Println("Unknown subType", subInfo.Type)
			return
		}
	}

	lock.Lock()
	defer lock.Unlock()
	subs, ok := (*subMap)[subInfo.Key]

	if !ok {
		subs = []types.Client{}
	}
	// if slices.Contains(subs, client) {
	// 	return
	// }

	(*subMap)[subInfo.Key] = append(subs, client)
}

func (cm *clientManager) RemoveSubscription(subInfo types.Subscription, client types.Client, removeAll bool) {
	var subMap *map[types.SubId][]types.Client
	var lock *sync.Mutex

	switch subInfo.Type {
	case types.FolderSubscribe:
		{
			subMap = &cm.folderSubs
			lock = cm.folderMu
		}
	case types.TaskSubscribe:
		{
			subMap = &cm.taskSubs
			lock = cm.taskMu
		}
	default:
		{
			util.Error.Println("Unknown subType", string(subInfo.Type))
			return
		}
	}

	lock.Lock()
	defer lock.Unlock()
	subs, ok := (*subMap)[subInfo.Key]
	if !ok {
		util.Warning.Println("Tried to unsubscribe from non-existent key", string(subInfo.Key))
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
	(*subMap)[subInfo.Key] = subs
}
