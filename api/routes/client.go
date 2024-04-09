package routes

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"slices"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func (c *Client) GetClientId() clientId {
	return c.connId
}

func (c *Client) GetShortId() clientId {
	return c.connId[28:]
}

func (c *Client) SetUser(user types.User) {
	c.user = user
}

func (c *Client) Username() types.Username {
	return c.user.GetUsername()
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
func (c *Client) Subscribe(subType subType, key subId, meta subMeta) (complete bool, results map[string]any) {
	var sub subscription

	switch subType {
	case SubFolder:
		{
			if key == "" {
				err := fmt.Errorf("cannot subscribe with empty folder id")
				c.Error(err)
				return
			}
			fileId := types.FileId(key)
			var folder types.WeblensFile
			if fileId == "external" {
				folder = dataStore.GetExternalDir()
			} else {
				folder = dataStore.FsTreeGet(fileId)
			}
			acc := dataStore.NewAccessMeta(c.user.GetUsername()).SetRequestMode(dataStore.FileSubscribeRequest)
			if folder == nil {
				err := fmt.Errorf("could not find folder with ID %s", key)
				c.Error(err)
				return
			} else if !dataStore.CanAccessFile(folder, acc) {
				err := fmt.Errorf("could not find folder with ID %s", key)
				c.Error(err)

				// dont tell the client they don't have access instead of *actually* not found
				c.err("User does not have access to ", key)
				return
			}

			sub = subscription{Type: subType, Key: key}
			c.PushFileUpdate(folder)

			// Subscribe to tasks on children in this folder
			for _, ch := range folder.GetChildren() {
				for _, t := range ch.GetTasks() {
					c.Subscribe(SubTask, subId(t.TaskId()), nil)
				}
			}
		}
	case SubTask:
		{
			task := dataProcess.GetTask(types.TaskId(key))
			if task == nil {
				err := fmt.Errorf("could not find task with ID %s", key)
				util.ErrTrace(err)
				c.Error(err)
				return
			}

			complete, _ = task.Status()
			if meta != nil {
				results = task.GetResult(meta.Meta(SubTask).(taskSubMetadata).ResultKeys()...)
			}
			if complete || slices.IndexFunc(c.subscriptions, func(s subscription) bool { return s.Key == key }) != -1 {
				return
			}

			sub = subscription{Type: subType, Key: key}
		}
	default:
		{
			err := fmt.Errorf("unknown subscription type %s", subType)
			util.ErrTrace(err)
			c.Error(err)
			return
		}
	}
	c.debug("Subscribed to", subType, key)

	c.mu.Lock()
	c.subscriptions = append(c.subscriptions, sub)
	c.mu.Unlock()
	cmInstance.AddSubscription(sub, c)

	return
}

func (c *Client) Unsubscribe(key subId) {
	c.mu.Lock()
	subIndex := slices.IndexFunc(c.subscriptions, func(s subscription) bool { return s.Key == key })
	if subIndex == -1 {
		c.mu.Unlock()
		return
	}
	var subToRemove subscription
	c.subscriptions, subToRemove = util.Yoink(c.subscriptions, subIndex)
	c.mu.Unlock()

	cmInstance.RemoveSubscription(subToRemove, c, false)

	c.debug("Unsubscribed from", subToRemove.Type, subToRemove.Key)
}

func (c *Client) Send(messageStatus string, key subId, content []wsM) {
	msg := wsResponse{MessageStatus: messageStatus, SubscribeKey: key, Content: content}
	// c.debug(fmt.Sprintf("Sending %s %s", messageStatus, key))
	c.writeToClient(msg)
}

func (c *Client) Error(err error) {
	switch err.(type) {
	case dataStore.WeblensFileError:
		c.err(err)
	default:
		c.errTrace(err)
	}

	msg := wsResponse{MessageStatus: "error", Error: err.Error()}
	c.writeToClient(msg)
}

func (c *Client) PushFileUpdate(updatedFile types.WeblensFile) {
	acc := dataStore.NewAccessMeta(c.user.GetUsername()).SetRequestMode(dataStore.WebsocketFileUpdate)
	fileInfo, err := updatedFile.FormatFileInfo(acc)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	c.Send("file_updated", subId(updatedFile.Id()), []wsM{{"fileInfo": fileInfo}})
}

func (c *Client) writeToClient(msg wsResponse) {
	if c != nil {
		c.mu.Lock()
		err := c.conn.WriteJSON(msg)
		c.mu.Unlock()
		if err != nil {
			c.errTrace(err)
		}
	}
}

func (c *Client) clientMsgFormat(msg ...any) string {
	return fmt.Sprintf("| %s %s | %s", c.GetShortId(), "("+c.user.GetUsername()+")", fmt.Sprintln(msg...))
}

func (c *Client) log(msg ...any) {
	util.WsInfo.Printf(c.clientMsgFormat(msg...))
}

func (c *Client) err(msg ...any) {
	_, file, line, _ := runtime.Caller(2)
	msg = []any{any(fmt.Sprintf("%s:%d:", file, line)), msg}
	util.WsError.Printf(c.clientMsgFormat(msg...))
}

func (c *Client) errTrace(msg ...any) {
	_, file, line, _ := runtime.Caller(2)
	msg = []any{any(fmt.Sprintf("%s:%d:", file, line)), msg}
	util.WsError.Printf(c.clientMsgFormat(msg...), string(debug.Stack()))
}

func (c *Client) debug(msg ...any) {
	util.WsDebug.Printf(c.clientMsgFormat(msg...))
}
