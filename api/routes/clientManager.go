package routes

import (
	"path/filepath"
	"slices"
	"sync"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var cmInstance clientManager

func VerifyClientManager() *clientManager {
	if cmInstance.clientMap == nil {
		cmInstance.clientMap = map[string]*Client{}
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

func (cm *clientManager) ClientConnect(conn *websocket.Conn, username string) *Client {
	cm = VerifyClientManager()
	connectionId := uuid.New().String()
	newClient := Client{connId: connectionId, conn: conn, username: username}
	cm.clientMu.Lock()
	cm.clientMap[connectionId] = &newClient
	cm.clientMu.Unlock()
	newClient.log("Connected", newClient.Username())
	return &newClient
}

func (cm clientManager) ClientDisconnect(c *Client) {
	for _, s := range c.subscriptions {
		cm.RemoveSubscription(s, c)
	}

	cm.clientMu.Lock()
	delete(cm.clientMap, c.GetClientId())
	cm.clientMu.Unlock()
}

func (cm clientManager) GetClient(clientId string) *Client {
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
	default:
		util.Error.Println("Unknown subscriber type", st)
	}

	return
}

func (cm clientManager) Broadcast(broadcastType subType, broadcastKey subId, messageStatus string, content []map[string]any) {
	if broadcastKey == "" {
		util.Error.Println("Trying to broadcast on empty key")
		return
	}
	defer util.RecoverPanic("Panic caught while broadcasting: %v")

	msg := wsResponse{MessageStatus: messageStatus, SubscribeKey: broadcastKey, Content: content}

	clients := cmInstance.GetSubscribers(subType(broadcastType), subId(broadcastKey))

	if len(clients) != 0 {
		for _, c := range clients {
			c._writeToClient(msg)
		}
	} else {
		// Although debug is our "verbose" mode, this one is *really* annoying, so it's disabled unless needed.
		// util.Debug.Println("No subscribers to", dest.Type, dest.Key)
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
	if slices.Contains(subs, client) {
		return
	}

	(*subMap)[subInfo.Key] = append(subs, client)
}

func (cm *clientManager) RemoveSubscription(subInfo subscription, client *Client) {
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
	subs = util.Filter(subs, func(c *Client) bool { return c.connId != client.connId })
	(*subMap)[subInfo.Key] = subs
}

func (c unbufferedCaster) PushFileCreate(newFile *dataStore.WeblensFile) {
	if !c.enabled {
		return
	}
	fileInfo, err := newFile.FormatFileInfo()
	if err != nil {
		util.DisplayError(err)
		return
	}

	cmInstance.Broadcast("folder", subId(newFile.GetParent().Id()), "file_created", []map[string]any{gin.H{"fileInfo": fileInfo}})
}

func (c unbufferedCaster) PushFileUpdate(updatedFile *dataStore.WeblensFile) {
	if !c.enabled {
		return
	}

	if dataStore.IsSystemDir(updatedFile) {
		return
	}

	fileInfo, err := updatedFile.FormatFileInfo()
	if err != nil {
		util.DisplayError(err)
		return
	}

	cmInstance.Broadcast("folder", subId(updatedFile.Id()), "file_updated", []map[string]any{gin.H{"fileInfo": fileInfo}})

	if dataStore.IsSystemDir(updatedFile.GetParent()) {
		return
	}
	cmInstance.Broadcast("folder", subId(updatedFile.GetParent().Id()), "file_updated", []map[string]any{gin.H{"fileInfo": fileInfo}})
}

func (c unbufferedCaster) PushFileMove(preMoveFile *dataStore.WeblensFile, postMoveFile *dataStore.WeblensFile) {
	if !c.enabled {
		return
	}

	if filepath.Dir(preMoveFile.String()) == filepath.Dir(postMoveFile.String()) {
		util.Error.Println("This should've been a rename")
		return
	}

	c.PushFileCreate(postMoveFile)
	c.PushFileDelete(preMoveFile)
}

func (c unbufferedCaster) PushFileDelete(deletedFile *dataStore.WeblensFile) {
	if !c.enabled {
		return
	}

	content := []map[string]any{gin.H{"fileId": deletedFile.Id()}}
	cmInstance.Broadcast("folder", subId(deletedFile.GetParent().Id()), "file_deleted", content)
}
