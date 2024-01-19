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
		cmInstance.clientMap = &map[string]*Client{}
		cmInstance.clientMu = &sync.Mutex{}
	}
	if cmInstance.folderSubs == nil {
		cmInstance.folderSubs = &map[subId][]*Client{}
		cmInstance.folderMu = &sync.Mutex{}
	}
	if cmInstance.taskSubs == nil {
		cmInstance.taskSubs = &map[subId][]*Client{}
		cmInstance.taskMu = &sync.Mutex{}
	}

	return &cmInstance
}

func (cm *clientManager) ClientConnect(conn *websocket.Conn) *Client {
	cm = VerifyClientManager()
	connectionId := uuid.New().String()
	newClient := Client{connId: connectionId, conn: conn}
	cm.clientMu.Lock()
	(*cm.clientMap)[connectionId] = &newClient
	cm.clientMu.Unlock()
	return &newClient
}

func (cm clientManager) ClientDisconnect(c *Client) {
	for _, s := range c.subscriptions {
		cm.RemoveSubscription(s, c)
	}

	cm.clientMu.Lock()
	delete((*cm.clientMap), c.GetClientId())
	cm.clientMu.Unlock()
}

func (cm clientManager) Broadcast(broadcastType subType, broadcastKey subId, messageStatus string, content any) {
	if broadcastKey == "" {
		util.Error.Println("Trying to broadcast on empty key")
		return
	}
	defer util.RecoverPanic("Panic caught while broadcasting: %v")

	msg := wsResponse{MessageStatus: messageStatus, SubscribeKey: broadcastKey, Content: content}
	dest := subscription{Type: subType(broadcastType), Key: subId(broadcastKey)}

	var allClients []*Client

	switch dest.Type {
	case Folder:
		{
			allClients = (*cm.folderSubs)[dest.Key]
		}
	case Task:
		{
			allClients = (*cm.taskSubs)[dest.Key]
		}
	}

	if len(allClients) != 0 {
		for _, c := range allClients {
			// util.Debug.Printf("Broadcasting %s for %v", msg.MessageStatus, msg.Content)
			c._writeToClient(msg)
		}
	} else {
		// util.Debug.Println("No subscribers to", dest.Type, dest.Key)
	}
}

func (cm clientManager) AddSubscription(subInfo subscription, client *Client) {
	var subMap *map[subId][]*Client
	var lock *sync.Mutex

	switch subInfo.Type {
	case Folder:
		{
			subMap = cm.folderSubs
			lock = cm.folderMu
		}
	case Task:
		{
			subMap = cm.taskSubs
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
	// util.Debug.Println("New subscriptions", (*subMap)[subInfo.Key])
}

func (cm *clientManager) RemoveSubscription(subInfo subscription, client *Client) {
	var subMap *map[subId][]*Client
	var lock *sync.Mutex

	switch subInfo.Type {
	case Folder:
		{
			subMap = cm.folderSubs
			lock = cm.folderMu
		}
	case Task:
		{
			subMap = cm.taskSubs
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

	// util.Debug.Println("Removed subscription")
}

func (c caster) PushTaskUpdate(taskId, status string, result any) {
	if !c.enabled {
		return
	}
	cmInstance.Broadcast("task", subId(taskId), status, result)
}

func (c caster) PushFileCreate(newFile *dataStore.WeblensFile) {
	if !c.enabled {
		return
	}
	fileInfo, err := newFile.FormatFileInfo()
	if err != nil {
		util.DisplayError(err)
		return
	}

	// util.Debug.Printf("Broadcasting create %s", newFile.String())
	cmInstance.Broadcast("folder", subId(newFile.GetParent().Id()), "file_created", map[string]any{"fileInfo": fileInfo})
}

func (c caster) PushFileUpdate(updatedFile *dataStore.WeblensFile) {
	if !c.enabled {
		return
	}

	fileInfo, err := updatedFile.FormatFileInfo()
	if err != nil {
		util.DisplayError(err)
		return
	}
	// util.Debug.Println("Broadcasting update", updatedFile.String())
	cmInstance.Broadcast("folder", subId(updatedFile.Id()), "file_updated", gin.H{"fileInfo": fileInfo})

	if updatedFile.GetParent().Id() == "0" {
		return
	}
	cmInstance.Broadcast("folder", subId(updatedFile.GetParent().Id()), "file_updated", gin.H{"fileInfo": fileInfo})
}

func (c caster) PushFileMove(preMoveFile *dataStore.WeblensFile, postMoveFile *dataStore.WeblensFile) {
	if !c.enabled {
		return
	}

	if filepath.Dir(preMoveFile.String()) == filepath.Dir(postMoveFile.String()) {
		// This should've been a "rename"
		util.Error.Println("This should've been a rename")
		return
	}

	// util.Debug.Printf("Broadcasting move %s -> %s", preMoveFile.String(), postMoveFile.String())
	c.PushFileCreate(postMoveFile)
	c.PushFileDelete(preMoveFile)
}

func (c caster) PushFileDelete(deletedFile *dataStore.WeblensFile) {
	if !c.enabled {
		return
	}

	content := gin.H{"fileId": deletedFile.Id()}
	cmInstance.Broadcast("folder", subId(deletedFile.GetParent().Id()), "file_deleted", content)
}
