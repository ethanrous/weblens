package routes

import (
	"slices"
	"sync"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var cmInstance clientManager

func VerifyClientManager() *clientManager {
	if cmInstance.clientMap == nil {
		cmInstance.clientMap = map[clientId]*Client{}
		cmInstance.clientMu = &sync.Mutex{}
	}
	if cmInstance.folderSubs == nil {
		cmInstance.folderSubs = map[subId][]*Client{}
		cmInstance.folderMu = &sync.Mutex{}
	}
	if cmInstance.taskSubs == nil {
		cmInstance.taskSubs = map[subId][]*Client{}
		cmInstance.taskMu = &sync.Mutex{}
	}

	return &cmInstance
}

func (cm *clientManager) ClientConnect(conn *websocket.Conn, user types.User) *Client {
	cm = VerifyClientManager()
	connectionId := clientId(uuid.New().String())
	newClient := Client{connId: connectionId, conn: conn, user: user}
	cm.clientMu.Lock()
	cm.clientMap[connectionId] = &newClient
	cm.clientMu.Unlock()
	newClient.log("Connected")
	return &newClient
}

func (cm clientManager) ClientDisconnect(c *Client) {
	for _, s := range c.subscriptions {
		cm.RemoveSubscription(s, c, true)
	}

	cm.clientMu.Lock()
	delete(cm.clientMap, c.GetClientId())
	cm.clientMu.Unlock()
}

func (cm clientManager) GetClient(clientId clientId) *Client {
	cm.clientMu.Lock()
	defer cm.clientMu.Unlock()
	return cm.clientMap[clientId]
}

func (cm clientManager) GetSubscribers(st subType, key subId) (clients []*Client) {

	switch st {
	case SubFolder:
		{
			clients = cm.folderSubs[key]
		}
	case SubTask:
		{
			clients = cm.taskSubs[key]
		}
	case SubUser:
		{
			allClients := util.MapToSlicePure(cm.clientMap)
			clients = util.Filter(allClients, func(c *Client) bool { return subId(c.user.GetUsername()) == key })
		}
	default:
		util.Error.Println("Unknown subscriber type", st)
	}

	return
}

func (cm clientManager) Broadcast(broadcastType subType, broadcastKey subId, messageStatus string, content []wsM) {
	if broadcastKey == "" {
		util.Error.Println("Trying to broadcast on empty key")
		return
	}
	defer util.RecoverPanic("Panic caught while broadcasting: %v")

	msg := wsResponse{MessageStatus: messageStatus, SubscribeKey: broadcastKey, Content: content}

	clients := cmInstance.GetSubscribers(subType(broadcastType), subId(broadcastKey))

	if len(clients) != 0 {
		for _, c := range clients {
			c.writeToClient(msg)
		}
	} else {
		// Although debug is our "verbose" mode, this one is *really* annoying, so it's disabled unless needed.
		// util.Debug.Println("No subscribers to", dest.Type, dest.Key)
		return
	}
}

func (cm *clientManager) AddSubscription(subInfo subscription, client *Client) {
	var subMap *map[subId][]*Client
	var lock *sync.Mutex

	switch subInfo.Type {
	case SubFolder:
		{
			subMap = &cm.folderSubs
			lock = cm.folderMu
		}
	case SubTask:
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
		subs = []*Client{}
	}
	// if slices.Contains(subs, client) {
	// 	return
	// }

	(*subMap)[subInfo.Key] = append(subs, client)
}

func (cm *clientManager) RemoveSubscription(subInfo subscription, client *Client, removeAll bool) {
	var subMap *map[subId][]*Client
	var lock *sync.Mutex

	switch subInfo.Type {
	case SubFolder:
		{
			subMap = &cm.folderSubs
			lock = cm.folderMu
		}
	case SubTask:
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
		util.Warning.Println("Tried to unsubscribe from non-existant key", string(subInfo.Key))
		return
	}
	if removeAll {
		subs = util.Filter(subs, func(c *Client) bool { return c.connId != client.connId })
	} else {
		index := slices.IndexFunc(subs, func(c *Client) bool { return c.connId == client.connId })
		if index != -1 {
			subs = util.Banish(subs, index)
		}
	}
	(*subMap)[subInfo.Key] = subs
}
