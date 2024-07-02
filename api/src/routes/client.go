package routes

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"slices"
	"sync"

	"github.com/ethrousseau/weblens/api/dataProcess"
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
	types.SERV.ClientManager.ClientDisconnect(c)

	c.mu.Lock()
	err := c.conn.Close()
	if err != nil {
		util.ShowErr(err)
		return
	}
	c.mu.Unlock()
	c.log("Disconnected")
}

// Subscribe links a websocket connection to a "key" that can be broadcast to later if
// relevant updates should be communicated
//
// Returns "true" and the results at meta.LookingFor if the task is completed, false and nil otherwise.
// Subscriptions to types that represent ongoing events like "folder" never return truthy completed
func (c *client) Subscribe(
	key types.SubId, action types.WsAction, acc types.AccessMeta,
	ft types.FileTree,
) (
	complete bool,
	results map[string]any,
) {
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
				folder = ft.Get("EXTERNAL")
			} else {
				folder = ft.Get(fileId)
			}

			acc.SetRequestMode(dataStore.FileSubscribeRequest)

			if folder == nil {
				err := fmt.Errorf("failed to find folder to subscribe to: %s", key)
				c.Error(err)
				return
			} else if !acc.CanAccessFile(folder) {
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
			task := types.SERV.TaskDispatcher.GetWorkerPool().GetTask(types.TaskId(key))
			if task == nil {
				err := fmt.Errorf("could not find task with ID %s", key)
				util.ErrTrace(err)
				c.Error(err)
				return
			}

			complete, _ = task.Status()
			results = task.GetResults()

			c.mu.Lock()
			if complete || slices.IndexFunc(
				c.subscriptions, func(s types.Subscription) bool { return s.Key == key },
			) != -1 {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()

			sub = types.Subscription{Type: types.TaskSubscribe, Key: key}

			c.PushTaskUpdate(task, dataProcess.TaskCreatedEvent, task.GetMeta().FormatToResult())
		}
	case types.PoolSubscribe:
		{
			pool := types.SERV.WorkerPool.GetTaskPool(types.TaskId(key))
			if pool == nil {
				c.Error(types.NewWeblensError(fmt.Sprintf("Could not find pool with id %s", key)))
				return
			} else if pool.IsGlobal() {
				c.Error(types.NewWeblensError("Trying to subscribe to global pool"))
				return
			}

			util.Debug.Printf("%s subscribed to pool [%s]", c.user.GetUsername(), pool.ID())
			sub = types.Subscription{Type: types.TaskSubscribe, Key: key}

			c.PushPoolUpdate(
				pool, dataProcess.PoolCreatedEvent, types.TaskResult{
					"createdBy": pool.CreatedInTask().
						TaskId(),
				},
			)
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
	types.SERV.ClientManager.AddSubscription(sub, c)

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

	types.SERV.ClientManager.RemoveSubscription(subToRemove, c, false)

	c.debug("Unsubscribed from", subToRemove.Type, subToRemove.Key)
}

func (c *client) Error(err error) {
	var weblensError types.WeblensError
	ok := errors.As(err, &weblensError)

	var msg types.WsResponseInfo
	switch ok {
	case true:
		c.err(err)
		msg = types.WsResponseInfo{EventTag: "error", Error: err.Error()}
	case false:
		c.errTrace(err)
		msg = types.WsResponseInfo{EventTag: "error", Error: "Masked unexpected server error"}
	}

	c.Send(msg)
}

func (c *client) PushFileUpdate(updatedFile types.WeblensFile) {
	acc := dataStore.NewAccessMeta(c.user).SetRequestMode(dataStore.WebsocketFileUpdate)
	fileInfo, err := updatedFile.FormatFileInfo(acc)
	if err != nil {
		util.ErrTrace(err)
		return
	}
	msg := types.WsResponseInfo{
		EventTag:      "file_updated",
		SubscribeKey:  types.SubId(updatedFile.ID()),
		Content:       types.WsC{"fileInfo": fileInfo},
		BroadcastType: types.TaskSubscribe,
	}

	c.Send(msg)
}

func (c *client) PushTaskUpdate(task types.Task, event types.TaskEvent, result types.TaskResult) {
	msg := types.WsResponseInfo{
		EventTag:      string(event),
		SubscribeKey:  types.SubId(task.TaskId()),
		Content:       types.WsC(result),
		TaskType:      task.TaskType(),
		BroadcastType: types.TaskSubscribe,
	}

	c.Send(msg)
	// c.Send(string(event), types.SubId(taskId), []types.WsC{types.WsC(result)})
}

func (c *client) PushPoolUpdate(pool types.TaskPool, event types.TaskEvent, result types.TaskResult) {
	if pool.IsGlobal() {
		util.Warning.Println("Not pushing update on global pool")
		return
	}

	msg := types.WsResponseInfo{
		EventTag:      string(event),
		SubscribeKey:  types.SubId(pool.ID()),
		Content:       types.WsC(result),
		TaskType:      pool.CreatedInTask().TaskType(),
		BroadcastType: types.TaskSubscribe,
	}

	c.Send(msg)
	// c.Send(string(event), types.SubId(taskId), []types.WsC{types.WsC(result)})
}

func (c *client) GetSubscriptions() []types.Subscription {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.subscriptions[0:len(c.subscriptions)]
}

func (c *client) Send(msg types.WsResponseInfo) {
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
