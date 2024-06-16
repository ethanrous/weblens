package routes

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"slices"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type client struct {
	Active        bool
	connId        types.ClientId
	conn          *websocket.Conn
	mu            sync.Mutex
	subscriptions []types.Subscription
	user          types.User
}

func (c *client) GetClientId() types.ClientId {
	return c.connId
}

func (c *client) GetShortId() types.ClientId {
	return c.connId[28:]
}

func (c *client) SetUser(user types.User) {
	c.user = user
}

func (c *client) GetUser() types.User {
	return c.user
}

func (c *client) Disconnect() {
	c.Active = false
	rc.ClientManager.ClientDisconnect(c)

	c.mu.Lock()
	c.conn.Close()
	c.mu.Unlock()
	c.log("Disconnected")
}

// Subscribe links a websocket connection to a "key" that can be broadcast to later if
// relevant updates should be communicated
//
// Returns "true" and the results at meta.LookingFor if the task is completed, false and nil otherwise.
// Subscriptions to types that represent ongoing events like "folder" never return truthy completed
func (c *client) Subscribe(key types.SubId, action types.WsAction, acc types.AccessMeta,
	ft types.FileTree) (complete bool,
	results map[string]any) {
	var sub types.Subscription

	switch action {
	case types.FolderSubscribe:
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
				folder = ft.Get(fileId)
			}

			acc.SetRequestMode(dataStore.FileSubscribeRequest)

			if folder == nil {
				err := fmt.Errorf("failed to find folder to subscribe to: %s", key)
				c.Error(err)
				return
			} else if !dataStore.CanAccessFile(folder, acc) {
				err := fmt.Errorf("failed to find folder to subscribe to: %s", key)
				c.Error(err)

				// don't tell the client they don't have access instead of *actually* not found
				c.err("User does not have access to", key)
				return
			}

			sub = types.Subscription{Type: types.FolderSubscribe, Key: key}
			c.PushFileUpdate(folder)

			// Subscribe to task on this folder
			if t := folder.GetTask(); t != nil {
				c.Subscribe(types.SubId(t.TaskId()), types.TaskSubscribe, acc, ft)
			}
		}
	case types.TaskSubscribe:
		{
			task := rc.TaskDispatcher.GetWorkerPool().GetTask(types.TaskId(key))
			if task == nil {
				err := fmt.Errorf("could not find task with ID %s", key)
				util.ErrTrace(err)
				c.Error(err)
				return
			}

			complete, _ = task.Status()

			results = task.GetResults()

			if complete || slices.IndexFunc(c.subscriptions, func(s types.Subscription) bool { return s.Key == key }) != -1 {
				return
			}

			sub = types.Subscription{Type: types.TaskSubscribe, Key: key}
		}
	default:
		{
			err := fmt.Errorf("unknown subscription type %s", action)
			util.ErrTrace(err)
			c.Error(err)
			return
		}
	}
	c.debug("Subscribed to", action, key)

	c.mu.Lock()
	c.subscriptions = append(c.subscriptions, sub)
	c.mu.Unlock()
	rc.ClientManager.AddSubscription(sub, c)

	return
}

func (c *client) Unsubscribe(key types.SubId) {
	c.mu.Lock()
	subIndex := slices.IndexFunc(c.subscriptions, func(s types.Subscription) bool { return s.Key == key })
	if subIndex == -1 {
		c.mu.Unlock()
		return
	}
	var subToRemove types.Subscription
	c.subscriptions, subToRemove = util.Yoink(c.subscriptions, subIndex)
	c.mu.Unlock()

	rc.ClientManager.RemoveSubscription(subToRemove, c, false)

	c.debug("Unsubscribed from", subToRemove.Type, subToRemove.Key)
}

func (c *client) Send(eventTag string, key types.SubId, content []types.WsMsg) {
	msg := wsResponse{EventTag: eventTag, SubscribeKey: key, Content: content}
	c.writeToClient(msg)
}

func (c *client) Error(err error) {
	_, ok := err.(types.WeblensError)
	var msg wsResponse
	switch ok {
	case true:
		c.err(err)
		msg = wsResponse{EventTag: "error", Error: err.Error()}
	case false:
		c.errTrace(err)
		msg = wsResponse{EventTag: "error", Error: "Masked unexpected server error"}
	}

	c.writeToClient(msg)
}

func (c *client) PushFileUpdate(updatedFile types.WeblensFile) {
	acc := dataStore.NewAccessMeta(c.user, updatedFile.GetTree()).SetRequestMode(dataStore.WebsocketFileUpdate)
	fileInfo, err := updatedFile.FormatFileInfo(acc, rc.MediaRepo)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	c.Send("file_updated", types.SubId(updatedFile.ID()), []types.WsMsg{{"fileInfo": fileInfo}})
}

func (c *client) PushTaskUpdate(taskId types.TaskId, event types.TaskEvent, result types.TaskResult) {
	c.Send(string(event), types.SubId(taskId), []types.WsMsg{types.WsMsg(result)})
}

func (c *client) GetSubscriptions() []types.Subscription {
	return c.subscriptions
}

func (c *client) writeToClient(msg wsResponse) {
	if c != nil {
		c.mu.Lock()
		defer c.mu.Unlock()
		err := c.conn.WriteJSON(msg)
		if err != nil {
			c.errTrace(err)
		}
	}
}

func (c *client) clientMsgFormat(msg ...any) string {
	return fmt.Sprintf("| %s %s | %s", c.GetShortId(), "("+c.user.GetUsername()+")", fmt.Sprintln(msg...))
}

func (c *client) log(msg ...any) {
	util.WsInfo.Printf(c.clientMsgFormat(msg...))
}

func (c *client) err(msg ...any) {
	_, file, line, _ := runtime.Caller(2)
	msg = []any{any(fmt.Sprintf("%s:%d:", file, line)), msg}
	util.WsError.Printf(c.clientMsgFormat(msg...))
}

func (c *client) errTrace(msg ...any) {
	_, file, line, _ := runtime.Caller(2)
	msg = []any{any(fmt.Sprintf("%s:%d:", file, line)), msg}
	util.WsError.Printf(c.clientMsgFormat(msg...), string(debug.Stack()))
}

func (c *client) debug(msg ...any) {
	util.WsDebug.Printf(c.clientMsgFormat(msg...))
}
