package routes

import (
	"sync"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type clientManager struct {
	// Key: connection id, value: client instance
	clientMap map[string]*Client
	clientMu  sync.Mutex

	// Key: subscription identifier, value: connection id
	// Use string -> bool map to take advantage of O(1) lookup time when removing clients
	// Bool represents if the subscription is `recursive`
	// {
	// 	"/path/to/subscribe/to": {
	// 		"clientId1": true
	// 		"clientId2": false
	// 	}
	// }
	folderSubscriptionMap map[string]map[string]bool
	taskSubscriptionMap   map[string]map[string]bool
	folderMu              sync.Mutex
	taskMu                sync.Mutex
}

var cmInstance clientManager

func verifyClientManager() *clientManager {
	if cmInstance.clientMap == nil {
		cmInstance.clientMap = map[string]*Client{}
	}
	if cmInstance.folderSubscriptionMap == nil {
		cmInstance.folderSubscriptionMap = map[string]map[string]bool{}
	}
	if cmInstance.taskSubscriptionMap == nil {
		cmInstance.taskSubscriptionMap = map[string]map[string]bool{}
	}

	return &cmInstance
}

func ClientConnect(conn *websocket.Conn) *Client {
	verifyClientManager()
	connectionId := uuid.New().String()
	newClient := Client{connId: connectionId, conn: conn}
	cmInstance.clientMu.Lock()
	cmInstance.clientMap[connectionId] = &newClient
	cmInstance.clientMu.Unlock()
	return &newClient
}

func Broadcast(broadcastType, broadcastKey, messageStatus string, content any) {
	if broadcastKey == "" {
		util.Error.Println("Trying to broadcast on empty key")
		return
	}

	// Just spawn a thread to handle the broadcast
	// This is a "best effort" method
	msg := WsResponse{MessageStatus: messageStatus, SubscribeKey: broadcastKey, Content: content}
	// util.Debug.Println("Casting", broadcastType, msg)
	_broadcast(broadcastType, broadcastKey, msg)
}

func _broadcast(broadcastType, key string, msg WsResponse) {
	defer util.RecoverPanic("Panic caught while broadcasting: ")
	var allClients map[string]bool

	switch broadcastType {
	case "folder":
		{
			allClients = cmInstance.folderSubscriptionMap[key]
		}
	case "task":
		{
			allClients = cmInstance.taskSubscriptionMap[key]
		}
	}

	if len(allClients) != 0 {
		for c := range allClients {
			cmInstance.clientMap[c]._writeToClient(msg)
		}
	} else {
		// util.Debug.Println("No subscribers to", key)
	}
}

func RemoveSubscription(s SubData, clientId string) {
	switch s.SubType {
	case "folder":
		{
			cmInstance.folderMu.Lock()
			defer cmInstance.folderMu.Unlock()
			delete(cmInstance.folderSubscriptionMap[s.SubKey], clientId)
		}
	case "task":
		{
			cmInstance.taskMu.Lock()
			defer cmInstance.taskMu.Unlock()
			delete(cmInstance.taskSubscriptionMap[s.SubKey], clientId)
		}
	}
}

func (c caster) PushTaskUpdate(taskId, status string, result any) {
	if !c.enabled {
		return
	}
	Broadcast("task", taskId, status, result)
}

func (c caster) PushItemCreate(newFile *dataStore.WeblensFileDescriptor) {
	if !c.enabled {
		return
	}
	fileInfo, err := newFile.FormatFileInfo()
	if err != nil {
		util.DisplayError(err)
		return
	}

	util.Debug.Printf("Broadcasting create %s to %s", newFile.String(), newFile.GetParent().Id())
	Broadcast("folder", newFile.GetParent().Id(), "file_created", map[string]any{"fileInfo": fileInfo})
}

func (c caster) PushItemUpdate(updatedFile *dataStore.WeblensFileDescriptor) {
	fileInfo, err := updatedFile.FormatFileInfo()
	if err != nil {
		util.DisplayError(err)
		return
	}
	Broadcast("folder", updatedFile.Id(), "file_updated", map[string]any{"fileInfo": fileInfo})

	util.Debug.Println("Broadcasting update", updatedFile.String())
	Broadcast("folder", updatedFile.GetParent().Id(), "file_updated", map[string]any{"fileInfo": fileInfo})
}

func (c caster) PushItemMove(preMoveFile *dataStore.WeblensFileDescriptor, postMoveFile *dataStore.WeblensFileDescriptor) {
	if !c.enabled {
		return
	}

	if preMoveFile.GetParent().String() == postMoveFile.GetParent().String() {
		// This should've been a "rename"
		util.Error.Println("This should've been a rename")
		return
	}

	util.Debug.Printf("Broadcasting move %s -> %s", preMoveFile.String(), postMoveFile.String())
	c.PushItemCreate(postMoveFile)
	c.PushItemDelete(preMoveFile)
}

func (c caster) PushItemDelete(deletedFile *dataStore.WeblensFileDescriptor) {
	if !c.enabled {
		return
	}
	content := map[string]any{"itemId": deletedFile.Id()}
	if deletedFile.Err() != nil {
		util.DisplayError(deletedFile.Err(), "Failed to get file Id while trying to push delete")
		return
	}

	util.Debug.Println("Broadcasting delete", deletedFile.String())
	Broadcast("folder", deletedFile.GetParent().Id(), "file_deleted", content)
}
