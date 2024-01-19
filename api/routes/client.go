package routes

import (
	"fmt"
	"slices"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func (c *Client) GetClientId() string {
	return c.connId
}

func (c *Client) GetShortId() string {
	return c.connId[28:]
}

func (c *Client) SetUser(username string) {
	c.username = username
}

func (c *Client) Username() string {
	return c.username
}

func (c *Client) Disconnect() {
	cmInstance.ClientDisconnect(c)

	c.mu.Lock()
	c.conn.Close()
	c.mu.Unlock()
	c.log("Disconnected")
}

// Link a websocket connection to a "key" that can be broadcasted to later if
// relevent updates should be communicated
//
// Returns "true" and the results at meta.LookingFor if the task is completed, false otherwise.
// Subscriptions to types that represent ongoing events like "folder" never return truthy completed
func (c *Client) Subscribe(subType subType, key subId, meta subMeta) (complete bool, result string) {
	var sub subscription

	switch subType {
	case Folder:
		{
			if key == "" {
				err := fmt.Errorf("cannot subscribe with empty folder id")
				util.DisplayError(err)
				c.Error(err)
				return
			}
			folder := dataStore.FsTreeGet(string(key))
			if folder == nil {
				err := fmt.Errorf("could not find folder with ID %s", key)
				util.DisplayError(err)
				c.Error(err)
				return
			}
			sub = subscription{Type: subType, Key: key}
			c.PushFileUpdate(folder)
		}
	case Task:
		{
			task := dataProcess.GetTask(string(key))
			if task == nil {
				err := fmt.Errorf("could not find task with ID %s", key)
				util.DisplayError(err)
				c.Error(err)
				return
			}

			complete = task.Completed
			if meta != nil {
				result = task.Result(meta.Meta(Task).(taskSubMetadata).ResultKeys()[0])
			}
			if complete {
				return
			}

			sub = subscription{Type: subType, Key: key}
		}
	default:
		{
			err := fmt.Errorf("unknown subscription type %s", subType)
			util.DisplayError(err)
			c.Error(err)
			return
		}
	}

	c.mu.Lock()
	c.subscriptions = append(c.subscriptions, sub)
	c.mu.Unlock()
	cmInstance.AddSubscription(sub, c)
	return
}

func (c *Client) Unsubscribe(key subId) {
	c.mu.Lock()
	subIndex := slices.IndexFunc[[]subscription](c.subscriptions, func(s subscription) bool { return s.Key == key })
	if subIndex == -1 {
		c.mu.Unlock()
		return
	}
	var subToRemove subscription
	c.subscriptions, subToRemove = util.Yoink(c.subscriptions, subIndex)
	c.mu.Unlock()

	cmInstance.RemoveSubscription(subToRemove, c)
	// c.log(fmt.Sprintf("Unsubscribed from %s. Now: %v", key, c.subscriptions))
}

// func (c *Client) isSubscribed(key subId) bool {
// 	return slices.ContainsFunc(c.subscriptions, func(s subscription) bool { return s.Key == key })
// }

func (c *Client) log(msg ...any) {
	util.WsInfo.Printf("| %s | %s", c.GetShortId(), fmt.Sprintln(msg...))
}

func (c *Client) _writeToClient(msg wsResponse) {
	if c != nil {
		c.mu.Lock()
		c.conn.WriteJSON(msg)
		c.mu.Unlock()
	}
}

func (c *Client) Send(messageStatus string, key subId, content any) {
	msg := wsResponse{MessageStatus: messageStatus, SubscribeKey: key, Content: content}
	c.log(fmt.Sprintf("Sending %s %s", messageStatus, key))
	c._writeToClient(msg)
}

func (c *Client) Error(err error) {
	msg := wsResponse{MessageStatus: "error", Error: err.Error()}
	c.log(fmt.Sprintf("Error %v", msg))
	c._writeToClient(msg)
}

func (c *Client) PushFileUpdate(updatedFile *dataStore.WeblensFile) {
	// if !c.isSubscribed(subId(updatedFile.Id())) {
	// 	return
	// }

	fileInfo, err := updatedFile.FormatFileInfo()
	if err != nil {
		util.DisplayError(err)
		return
	}

	c.Send("file_updated", subId(updatedFile.Id()), gin.H{"fileInfo": fileInfo})
}
