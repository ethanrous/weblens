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
	clientMap map[types.ClientId]types.Client
	clientMu  *sync.Mutex

	// Key: subscription identifier, value: connection id
	// Use string -> bool map to take advantage of O(1) lookup time when removing clients
	// Bool represents if the subscription is `recursive`
	// {
	// 	"fileId": {
	// 		"clientId1": true
	// 		"clientId2": false
	// 	}
	// }
	folderSubs map[types.SubId][]types.Client
	taskSubs   map[types.SubId][]types.Client
	folderMu   *sync.Mutex
	taskMu     *sync.Mutex
}

func NewClientManager() types.ClientManager {
	return &clientManager{
		clientMap: map[types.ClientId]types.Client{},
		clientMu:  &sync.Mutex{},

		folderSubs: map[types.SubId][]types.Client{},
		taskSubs:   map[types.SubId][]types.Client{},

		folderMu: &sync.Mutex{},
		taskMu:   &sync.Mutex{},
	}
}

// func VerifyClientManager() *clientManager {
// 	if cmInstance.clientMap == nil {
// 		cmInstance.clientMap = map[clientId]*client{}
// 		cmInstance.clientMu = &sync.Mutex{}
// 	}
// 	if cmInstance.folderSubs == nil {
// 		cmInstance.folderSubs = map[types.SubId][]*client{}
// 		cmInstance.folderMu = &sync.Mutex{}
// 	}
// 	if cmInstance.taskSubs == nil {
// 		cmInstance.taskSubs = map[types.SubId][]*client{}
// 		cmInstance.taskMu = &sync.Mutex{}
// 	}
//
// 	return &cmInstance
// }

func (cm *clientManager) ClientConnect(conn *websocket.Conn, user types.User) types.Client {
	connectionId := types.ClientId(uuid.New().String())
	newClient := client{Active: true, connId: connectionId, conn: conn, user: user}

	cm.clientMu.Lock()
	cm.clientMap[connectionId] = &newClient
	cm.clientMu.Unlock()

	newClient.log("Connected")
	return &newClient
}

func (cm *clientManager) ClientDisconnect(c types.Client) {
	for _, s := range c.GetSubscriptions() {
		cm.RemoveSubscription(s, c, true)
	}

	cm.clientMu.Lock()
	delete(cm.clientMap, c.GetClientId())
	cm.clientMu.Unlock()
}

func (cm *clientManager) GetClient(clientId types.ClientId) types.Client {
	cm.clientMu.Lock()
	defer cm.clientMu.Unlock()
	return cm.clientMap[clientId]
}

func (cm *clientManager) GetSubscribers(st types.WsAction, key types.SubId) (clients []types.Client) {
	switch st {
	case types.FolderSubscribe:
		{
			clients = cm.folderSubs[key]
		}
	case types.TaskSubscribe:
		{
			clients = cm.taskSubs[key]
		}
	case types.SubUser:
		{
			allClients := util.MapToSlicePure(cm.clientMap)
			clients = util.Filter(
				allClients, func(c types.Client) bool {
					return types.SubId(c.GetUser().GetUsername()) == key
				},
			)
		}
	default:
		util.Error.Println("Unknown subscriber type", st)
	}

	return
}

func (cm *clientManager) Broadcast(
	broadcastType types.WsAction, broadcastKey types.SubId, eventTag string, content []types.WsMsg,
) {
	if broadcastKey == "" {
		util.Error.Println("Trying to broadcast on empty key")
		return
	}
	defer util.RecoverPanic("Panic caught while broadcasting")

	msg := wsResponse{EventTag: eventTag, SubscribeKey: broadcastKey, Content: content}

	clients := cm.GetSubscribers(broadcastType, broadcastKey)

	if len(clients) != 0 {
		for _, c := range clients {
			c.(*client).writeToClient(msg)
		}
	} else {
		// Although debug is our "verbose" mode, this one is *really* annoying, so it's disabled unless needed.
		// util.Debug.Println("No subscribers to", dest.Type, dest.Key)
		return
	}
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
