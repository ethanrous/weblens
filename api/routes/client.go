package routes

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ethrousseau/weblens/api/dataProcess"
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
	activeFolder string // This is the same as the key in the folder subscription map
	mu sync.Mutex
	subscriptions []SubData
	username string
}

func (c *Client) GetClientId() string {
	return c.connId
}

func (c *Client) SetUser(username string) {
	c.username = username
}

func (c *Client) Username() string {
	return c.username
}

func (c *Client) Disconnect() {
	cmInstance.folderMu.Lock()
	defer cmInstance.folderMu.Unlock()
	if (c.activeFolder != "") {
		delete(cmInstance.folderSubscriptionMap[c.activeFolder], c.connId)
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

func (c *Client) updateFolderSubscription(folderId string, recursive bool) {
	cmInstance.folderMu.Lock()
	defer cmInstance.folderMu.Unlock()

	if (c.activeFolder != "") {
		delete(cmInstance.folderSubscriptionMap[c.activeFolder], c.connId)
	}

	_, ok := cmInstance.folderSubscriptionMap[folderId]
	if ok {
		cmInstance.folderSubscriptionMap[folderId][c.connId] = recursive
	} else {
		cmInstance.folderSubscriptionMap[folderId] = map[string]bool{c.connId: recursive}
	}

	c.activeFolder = folderId
}

// Link a websocket connection to a "key" that can be broadcasted to later if
// relevent updates should be communicated
//
// Returns "true" and the results at meta.LookingFor if the task is completed, false otherwise.
// Subscriptions to types that represent ongoing events like "folder" never return truthy completed
func (c *Client) Subscribe(subType, subData any) (bool, string) {
	switch subType {
	case "folder": {
		var meta dataProcess.FolderSubMetadata = subData.(dataProcess.FolderSubMetadata)

		if meta.FolderId == "" {
			panic(fmt.Errorf("empty folder id while trying to subscribe"))
		}
		folder := dataStore.FsTreeGet(meta.FolderId)
		if folder == nil {
			c.Send("error", nil, errors.New("could not find folder to subscribe to"))
			return false, ""
		}
		c.updateFolderSubscription(folder.Id(), meta.Recursive)
	}
	case "task": {
		meta := subData.(dataProcess.TaskSubMetadata)

		task := dataProcess.GetTask(meta.TaskId)
		if task == nil {
			util.Warning.Println("Could not find task with ID ", meta.TaskId)
			return false, ""
		} else if task.Completed {
			return true, task.Result(meta.LookingFor[0])
		}

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
	msg := WsResponse{MessageStatus: messageStatus, Content: content, Error: err.Error()}
	util.Debug.Printf("Sending to client [ %s ]\n%v", c.GetClientId(), msg)
	go c._writeToClient(msg)
}