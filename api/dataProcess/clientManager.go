package dataProcess

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type clientManager struct {
	// Key: connection id, value: client instance
	clientMap map[string]*Client
	clientMu sync.Mutex

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
	taskSubscriptionMap map[string]map[string]bool
	folderMu sync.Mutex
	taskMu sync.Mutex
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
	// Just spawn a thread to handle the broadcast
	// This is a "best effort" method

	// util.Debug.Printf("Broadcast: [%s %s] -- %v", label, key, content)
	msg := WsResponse{MessageStatus: messageStatus, SubscribeKey: broadcastKey, Content: content, Error: nil}
	go _broadcast(broadcastType, broadcastKey, msg)
}

func _broadcast(broadcastType, key string, msg WsResponse) {
	defer util.RecoverPanic("Panic caught while broadcasting: ")
	var allClients map[string]bool = make(map[string]bool)

	switch broadcastType {
	case "folder": {
		tmpKey := key
		for {
			tmpClients := cmInstance.folderSubscriptionMap[tmpKey]
			for c := range tmpClients {
				if tmpKey == key || tmpClients[c] {
					allClients[c] = true
				}
			}
			if dataStore.GuaranteeRelativePath(tmpKey) == "/" || filepath.Dir(tmpKey) == tmpKey {
				break
			}
			tmpKey = filepath.Dir(tmpKey)
		}
	}
	case "task": {
		allClients = cmInstance.taskSubscriptionMap[key]
	}
	}

	if len(allClients) != 0 {
		for c := range allClients {
			cmInstance.clientMap[c]._writeToClient(msg)
		}
	}
	// else {
	// 	// _, file, line, _ := runtime.Caller(1)
	// 	// util.Debug.Printf("No subscribers to %s (from %s:%d)", key, file, line)
	// }
}

func RemoveSubscription(s SubData, clientId string) {
	switch s.SubType {
	case "folder": {
		cmInstance.folderMu.Lock()
		defer cmInstance.folderMu.Unlock()
		delete(cmInstance.folderSubscriptionMap[s.SubKey], clientId)
	}
	case "task": {
		cmInstance.taskMu.Lock()
		defer cmInstance.taskMu.Unlock()
		delete(cmInstance.taskSubscriptionMap[s.SubKey], clientId)
	}
	}
}

var updateDb = dataStore.NewDB("")

func PushItemCreate(file *dataStore.WeblensFileDescriptor) {
	PushItemUpdate(file, file)
}

func PushItemUpdate(preUpdateFile *dataStore.WeblensFileDescriptor, postUpdateFile *dataStore.WeblensFileDescriptor) {
	fileInfo, err := postUpdateFile.FormatFileInfo()
	if err != nil {
		util.DisplayError(err, "Failed to push item update for: ", postUpdateFile.String())
		return
	}

	if preUpdateFile.Id() == "" {
		panic(fmt.Errorf("tried to push update for item with empty id: %s -- %s", preUpdateFile, preUpdateFile.Err()))
	}

	Broadcast("folder", postUpdateFile.ParentFolderId, "item_update", map[string]any{"itemId": preUpdateFile.Id(), "updateInfo": fileInfo})
	updateDb.RedisCacheBust(postUpdateFile.ParentFolderId)

	// When a file changes directories, we need to alert both the old (here vv ) and new (above ^^) folder subscribers of the change
	if postUpdateFile.ParentFolderId != preUpdateFile.ParentFolderId {
		Broadcast("folder", preUpdateFile.ParentFolderId, "item_update", map[string]any{"itemId": preUpdateFile.Id(), "updateInfo": fileInfo})
		updateDb.RedisCacheBust(preUpdateFile.ParentFolderId)
	}
}

func PushItemDelete(file *dataStore.WeblensFileDescriptor) {
	content := map[string]any{"itemId": file.Id()}
	if file.Err() != nil {
		util.DisplayError(file.Err(), "Failed to get file Id while trying to push delete")
		return
	}
	Broadcast("folder", file.ParentFolderId, "item_deleted", content)
	updateDb.RedisCacheBust(file.ParentFolderId)
}