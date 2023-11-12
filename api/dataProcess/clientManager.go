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

	// Key: subscription identifier, value: connection id
	// Use string -> bool map to take advantage of O(1) lookup time when removing clients
	// Bool represents if the subscription is `recursive`
	// {
	// 	"/path/to/subscribe/to": {
	// 		"clientId1": true
	// 		"clientId2": false
	// 	}
	// }
	subscriptionMap map[string]map[string]bool
	mu sync.Mutex
}

type Client struct {
	connId string
	conn *websocket.Conn
	activeCtx string // This is the same as the key in the subscription map
	mu sync.Mutex
}

var cmInstance clientManager

func verifyClientManager() *clientManager {
	if cmInstance.clientMap == nil {
		cmInstance.clientMap = map[string]*Client{}
	}
	if cmInstance.subscriptionMap == nil {
		cmInstance.subscriptionMap = map[string]map[string]bool{}
	}

	return &cmInstance
}

func ClientConnect(conn *websocket.Conn) *Client {
	verifyClientManager()
	connectionId := uuid.New().String()
	newClient := Client{connId: connectionId, conn: conn}
	cmInstance.mu.Lock()
	cmInstance.clientMap[connectionId] = &newClient
	cmInstance.mu.Unlock()
	return &newClient
}

func Broadcast(key string, msg any) () {
	var allClients map[string]bool = make(map[string]bool)

	tmpKey := key
	for {
		tmpClients := cmInstance.subscriptionMap[tmpKey]
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
	if len(allClients) != 0 {
		for c := range allClients {
			go cmInstance.clientMap[c].writeToClient(msg)
		}
	} else {
		_, file, line, _ := runtime.Caller(1)
		util.Debug.Printf("No subscribers to %s (from %s:%d)", key, file, line)
	}
}

func (c *Client) GetClientId() string {
	return c.connId
}

func (c *Client) removeSubscription() {
	// Remove old subscription
	if c.activeCtx != "" {
		cmInstance.mu.Lock()
		delete(cmInstance.subscriptionMap[c.activeCtx], c.connId)
		cmInstance.mu.Unlock()
	}
}

func (c *Client) Disconnect() {
	c.removeSubscription()
	c.mu.Lock()
	cmInstance.clientMap[c.connId].conn.Close()
	c.mu.Unlock()

	cmInstance.mu.Lock()
	delete(cmInstance.clientMap, c.connId)
	cmInstance.mu.Unlock()
}

func (c *Client) Subscribe(subMeta SubscribeContent, username string) {
	c.removeSubscription()

	absPath := dataStore.GuaranteeUserAbsolutePath(subMeta.Path, username)
	util.Debug.Printf("Subscribing to %s (recursive: %t)", absPath, subMeta.Recursive)
	c.activeCtx = absPath

	cmInstance.mu.Lock()
	cmInstance.clientMap[c.connId] = c

	_, ok := cmInstance.subscriptionMap[absPath]

	if ok {
		cmInstance.subscriptionMap[absPath][c.connId] = subMeta.Recursive
	} else {
		cmInstance.subscriptionMap[absPath] = map[string]bool{c.connId: subMeta.Recursive}
	}
	cmInstance.mu.Unlock()
}

func (c *Client) Send(msg any) {
	c.conn.WriteJSON(msg)
}

func (c *Client) writeToClient(msg any) {
	c.mu.Lock()
	c.conn.WriteJSON(msg)
	c.mu.Unlock()
}