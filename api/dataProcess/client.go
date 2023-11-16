package dataProcess

import (
	"sync"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gorilla/websocket"
)

type SubData struct {
	SubType string
	SubKey string
}

type Client struct {
	connId string
	conn *websocket.Conn
	activePath string // This is the same as the key in the path subscription map
	mu sync.Mutex
	subscriptions []SubData
}

func (c *Client) GetClientId() string {
	return c.connId
}

func (c *Client) Disconnect() {
	cmInstance.pathMu.Lock()
	defer cmInstance.pathMu.Unlock()
	if (c.activePath != "") {
		delete(cmInstance.pathSubscriptionMap[c.activePath], c.connId)
	}

	for _, s := range c.subscriptions {
		RemoveSubscription(s, c.GetClientId())
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	cmInstance.clientMap[c.connId].conn.Close()

	cmInstance.clientMu.Lock()
	defer cmInstance.clientMu.Unlock()
	delete(cmInstance.clientMap, c.connId)
}

func (c *Client) updatePathSubscription(newPath string, recursive bool) {
	cmInstance.pathMu.Lock()
	defer cmInstance.pathMu.Unlock()

	if (c.activePath != "") {
		delete(cmInstance.pathSubscriptionMap[c.activePath], c.connId)
	}

	_, ok := cmInstance.pathSubscriptionMap[newPath]
	if ok {
		cmInstance.pathSubscriptionMap[newPath][c.connId] = recursive
	} else {
		cmInstance.pathSubscriptionMap[newPath] = map[string]bool{c.connId: recursive}
	}

	c.activePath = newPath
}

// Link a websocket connection to a "key" that can be broadcasted to later if
// relevent updates should be communicated
//
// Returns "true" and the results at meta.LookingFor if the task is completed, false otherwise
// subscriptions to ongoing events like "path" never return truthful completed
func (c *Client) Subscribe(subType, username string, subData any) (bool, string) {
	// c.removeSubscription(subMeta.Label)

	switch subType {
	case "path": {
		var meta PathSubMetadata = subData.(PathSubMetadata)
		absDirPath := dataStore.GuaranteeUserAbsolutePath(meta.DirPath, username)
		c.updatePathSubscription(absDirPath, meta.Recursive)
	}
	case "task": {
		var meta TaskSubMetadata = subData.(TaskSubMetadata)

		task := GetTask(meta.TaskId)
		if task == nil {
			util.Debug.Println("Could not find task with ID ", meta.TaskId)
			return false, ""
		} else if task.Completed{
			util.Debug.Println("Found completed task")
			return true, task.result[meta.LookingFor[0]]
		}

		util.Debug.Println("Subscribing to incomplete task", meta.TaskId)
		cmInstance.taskMu.Lock()
		defer cmInstance.taskMu.Unlock()
		_, ok := cmInstance.taskSubscriptionMap[meta.TaskId]
		if ok {
			cmInstance.taskSubscriptionMap[meta.TaskId][c.connId] = true
		} else {
			cmInstance.taskSubscriptionMap[meta.TaskId] = map[string]bool{c.connId: true}
		}
		c.subscriptions = append(c.subscriptions, SubData{SubType: "task", SubKey: meta.TaskId})
	}
	default: {
		util.Error.Printf("Recieved unknown subscription type: [%s] -- Raw metadata: %v", subType, subData)
	}
	}
	return false, ""
}

func (c *Client) _writeToClient(msg WsResponse) {
	if c != nil {
		c.mu.Lock()
		c.conn.WriteJSON(msg)
		c.mu.Unlock()
	}
}

func (c *Client) Send(messageStatus string, content any, err error) {
	msg := WsResponse{MessageStatus: messageStatus, Content: content, Error: err}
	go c._writeToClient(msg)
}