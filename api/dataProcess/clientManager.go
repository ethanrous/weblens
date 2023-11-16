package dataProcess

import (
	"path/filepath"
	"runtime"
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
	pathSubscriptionMap map[string]map[string]bool
	taskSubscriptionMap map[string]map[string]bool
	pathMu sync.Mutex
	taskMu sync.Mutex
}

var cmInstance clientManager

func verifyClientManager() *clientManager {
	if cmInstance.clientMap == nil {
		cmInstance.clientMap = map[string]*Client{}
	}
	if cmInstance.pathSubscriptionMap == nil {
		cmInstance.pathSubscriptionMap = map[string]map[string]bool{}
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
	msg := WsResponse{MessageStatus: messageStatus, Content: content, Error: nil}
	go _broadcast(broadcastType, broadcastKey, msg)
}

func _broadcast(broadcastType, key string, msg WsResponse) {
	defer func(){err := recover(); if err != nil {util.Debug.Println("Got error while broadcasting:", err)}}()
	var allClients map[string]bool = make(map[string]bool)

	switch broadcastType {
	case "path": {

		tmpKey := key
		for {
			tmpClients := cmInstance.pathSubscriptionMap[tmpKey]
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
	} else {
		_, file, line, _ := runtime.Caller(1)
		util.Debug.Printf("No subscribers to %s (from %s:%d)", key, file, line)
	}
}

func RemoveSubscription(s SubData, clientId string) {
	switch s.SubType {
	case "path": {
		cmInstance.pathMu.Lock()
		defer cmInstance.pathMu.Unlock()
		delete(cmInstance.pathSubscriptionMap[s.SubKey], clientId)
	}
	case "task": {
		cmInstance.taskMu.Lock()
		defer cmInstance.taskMu.Unlock()
		delete(cmInstance.taskSubscriptionMap[s.SubKey], clientId)
	}
	}
}

func PushItemUpdate(path, username string, db dataStore.Weblensdb) {
	fileInfo, _ := dataStore.FormatFileInfo(path, username, db)
	parentDir := filepath.Dir(path)
	Broadcast("path", parentDir, "item_update", fileInfo)
	db.RedisCacheBust(parentDir)
}

func PushItemDelete(path, hash, username string, db dataStore.Weblensdb) {
	relPath, _ := dataStore.GuaranteeUserRelativePath(path, username)
	content := struct { Path string `json:"path"`; Hash string `json:"hash"`} {Path: relPath, Hash: hash}
	parentDir := filepath.Dir(path)
	Broadcast("path", parentDir, "item_deleted", content)
	db.RedisCacheBust(parentDir)
}