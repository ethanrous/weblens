package routes

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"slices"
	"sync"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/util/wlog"
	"github.com/gorilla/websocket"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type clientConn struct {
	Active        bool
	connId        types.ClientId
	conn          *websocket.Conn
	mu            sync.Mutex
	subscriptions []types.Subscription
	user          types.User
	remote        types.Instance
}

func (cc *clientConn) IsOpen() bool {
	return cc.Active
}

func (cc *clientConn) GetClientId() types.ClientId {
	return cc.connId
}

func (cc *clientConn) ClientType() types.ClientType {
	return types.WebClient
}

func (cc *clientConn) GetShortId() types.ClientId {
	if cc.connId == "" {
		return ""
	}
	return cc.connId[28:]
}

func (cc *clientConn) GetUser() types.User {
	return cc.user
}

func (cc *clientConn) GetRemote() types.Instance {
	return cc.remote
}

func (cc *clientConn) ReadOne() (int, []byte, error) {
	return cc.conn.ReadMessage()
}

func (cc *clientConn) Disconnect() {
	cc.Active = false
	types.SERV.ClientManager.ClientDisconnect(cc)

	cc.mu.Lock()
	err := cc.conn.Close()
	if err != nil {
		wlog.ShowErr(err)
		return
	}
	cc.mu.Unlock()
	cc.log("Disconnected")
}

// Subscribe links a websocket connection to a "key" that can be broadcast to later if
// relevant updates should be communicated
//
// Returns "true" and the results at meta.LookingFor if the task is completed, false and nil otherwise.
// Subscriptions to types that represent ongoing events like "folder" never return truthy completed
func (cc *clientConn) Subscribe(
	key types.SubId, action types.WsAction, acc types.AccessMeta,
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
				cc.Error(err)
				return
			}
			fileId := types.FileId(key)
			var folder types.WeblensFile
			if fileId == "external" {
				folder = types.SERV.FileTree.Get("EXTERNAL")
			} else {
				folder = types.SERV.FileTree.Get(fileId)
			}

			acc.SetRequestMode(dataStore.FileSubscribeRequest)

			if folder == nil {
				err := types.WeblensErrorMsg(fmt.Sprint("failed to find folder to subscribe to: ", key))
				cc.Error(err)
				return
			} else if !acc.CanAccessFile(folder) {
				err := types.WeblensErrorMsg(fmt.Sprint("failed to find folder to subscribe to: ", key))
				cc.Error(err)

				// don't tell the clientConn they don't have access instead of *actually* not found
				cc.err("User does not have access to", key)
				return
			}

			sub = types.Subscription{Type: types.FolderSubscribe, Key: key}
			cc.PushFileUpdate(folder)

			// Subscribe to task on this folder
			if t := folder.GetTask(); t != nil {
				cc.Subscribe(types.SubId(t.TaskId()), types.TaskSubscribe, acc)
			}
		}
	case types.TaskSubscribe:
		{
			task := types.SERV.TaskDispatcher.GetWorkerPool().GetTask(types.TaskId(key))
			if task == nil {
				err := fmt.Errorf("could not find task with ID %s", key)
				cc.Error(err)
				return
			}

			complete, _ = task.Status()
			results = task.GetResults()

			cc.mu.Lock()
			if complete || slices.IndexFunc(
				cc.subscriptions, func(s types.Subscription) bool { return s.Key == key },
			) != -1 {
				cc.mu.Unlock()
				return
			}
			cc.mu.Unlock()

			sub = types.Subscription{Type: types.TaskSubscribe, Key: key}

			cc.PushTaskUpdate(task, dataProcess.TaskCreatedEvent, task.GetMeta().FormatToResult())
		}
	case types.PoolSubscribe:
		{
			pool := types.SERV.WorkerPool.GetTaskPool(types.TaskId(key))
			if pool == nil {
				cc.Error(types.NewWeblensError(fmt.Sprintf("Could not find pool with id %s", key)))
				return
			} else if pool.IsGlobal() {
				cc.Error(types.NewWeblensError("Trying to subscribe to global pool"))
				return
			}

			wlog.Debug.Printf("%s subscribed to pool [%s]", cc.user.GetUsername(), pool.ID())
			sub = types.Subscription{Type: types.TaskSubscribe, Key: key}

			cc.PushPoolUpdate(
				pool, dataProcess.PoolCreatedEvent, types.TaskResult{
					"createdBy": pool.CreatedInTask().
						TaskId(),
				},
			)
		}
	case types.TaskTypeSubscribe:
		{
			wlog.Debug.Printf("%s subscribed to task type [%s]", cc.user.GetUsername(), key)
			sub = types.Subscription{Type: types.TaskTypeSubscribe, Key: key}
		}
	default:
		{
			err := fmt.Errorf("unknown subscription type %s", action)
			wlog.ErrTrace(err)
			cc.Error(err)
			return
		}
	}
	cc.debug("Subscribed to", action, key)

	cc.mu.Lock()
	cc.subscriptions = append(cc.subscriptions, sub)
	cc.mu.Unlock()
	types.SERV.ClientManager.AddSubscription(sub, cc)

	return
}

func (cc *clientConn) Unsubscribe(key types.SubId) {
	cc.mu.Lock()
	subIndex := slices.IndexFunc(cc.subscriptions, func(s types.Subscription) bool { return s.Key == key })
	if subIndex == -1 {
		cc.mu.Unlock()
		return
	}
	var subToRemove types.Subscription
	cc.subscriptions, subToRemove = util.Yoink(cc.subscriptions, subIndex)
	cc.mu.Unlock()

	types.SERV.ClientManager.RemoveSubscription(subToRemove, cc, false)

	cc.debug("Unsubscribed from", subToRemove.Type, subToRemove.Key)
}

func (cc *clientConn) Error(err error) {
	var weblensError types.WeblensError
	ok := errors.As(err, &weblensError)

	var msg types.WsResponseInfo
	switch ok {
	case true:
		cc.err(err)
		msg = types.WsResponseInfo{EventTag: "error", Error: err.Error()}
	case false:
		cc.errTrace(err)
		msg = types.WsResponseInfo{EventTag: "error", Error: "Masked unexpected server error"}
	}

	cc.Send(msg)
}

func (cc *clientConn) PushWeblensEvent(eventTag string) {
	msg := types.WsResponseInfo{
		EventTag:      eventTag,
		SubscribeKey:  types.SubId("WEBLENS"),
		BroadcastType: types.ServerEvent,
	}

	cc.Send(msg)
}

func (cc *clientConn) PushFileUpdate(updatedFile types.WeblensFile) {
	acc := dataStore.NewAccessMeta(cc.user).SetRequestMode(dataStore.WebsocketFileUpdate)
	fileInfo, err := updatedFile.FormatFileInfo(acc)
	if err != nil {
		wlog.ErrTrace(err)
		return
	}
	msg := types.WsResponseInfo{
		EventTag:      "file_updated",
		SubscribeKey:  types.SubId(updatedFile.ID()),
		Content:       types.WsC{"fileInfo": fileInfo},
		BroadcastType: types.FolderSubscribe,
	}

	cc.Send(msg)
}

func (cc *clientConn) PushTaskUpdate(task types.Task, event types.TaskEvent, result types.TaskResult) {
	msg := types.WsResponseInfo{
		EventTag:      string(event),
		SubscribeKey:  types.SubId(task.TaskId()),
		Content:       types.WsC(result),
		TaskType:      task.TaskType(),
		BroadcastType: types.TaskSubscribe,
	}

	cc.Send(msg)
	// c.Send(string(event), types.SubId(taskId), []types.WsC{types.WsC(result)})
}

func (cc *clientConn) PushPoolUpdate(pool types.TaskPool, event types.TaskEvent, result types.TaskResult) {
	if pool.IsGlobal() {
		wlog.Warning.Println("Not pushing update on global pool")
		return
	}

	msg := types.WsResponseInfo{
		EventTag:      string(event),
		SubscribeKey:  types.SubId(pool.ID()),
		Content:       types.WsC(result),
		TaskType:      pool.CreatedInTask().TaskType(),
		BroadcastType: types.TaskSubscribe,
	}

	cc.Send(msg)
	// c.Send(string(event), types.SubId(taskId), []types.WsC{types.WsC(result)})
}

func (cc *clientConn) GetSubscriptions() []types.Subscription {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return cc.subscriptions[0:len(cc.subscriptions)]
}

func (cc *clientConn) Send(msg types.WsResponseInfo) {
	if cc != nil {
		cc.mu.Lock()
		defer cc.mu.Unlock()
		err := cc.conn.WriteJSON(msg)
		if err != nil {
			cc.errTrace(err)
		}
	}
}

func (cc *clientConn) clientMsgFormat(msg ...any) string {
	var clientName string
	if cc.GetUser() != nil {
		clientName = string(cc.GetUser().GetUsername())
	} else {
		clientName = cc.GetRemote().GetName()
	}
	return fmt.Sprintf("| %s (%s) | %s", cc.GetShortId(), clientName, fmt.Sprintln(msg...))
}

func (cc *clientConn) log(msg ...any) {
	wlog.WsInfo.Printf(cc.clientMsgFormat(msg...))
}

func (cc *clientConn) err(msg ...any) {
	_, file, line, _ := runtime.Caller(2)
	msg = []any{any(fmt.Sprintf("%s:%d:", file, line)), msg}
	wlog.WsError.Printf(cc.clientMsgFormat(msg...))
}

func (cc *clientConn) errTrace(msg ...any) {
	_, file, line, _ := runtime.Caller(2)
	msg = []any{any(fmt.Sprintf("%s:%d:", file, line)), msg}
	wlog.WsError.Printf(cc.clientMsgFormat(msg...), string(debug.Stack()))
}

func (cc *clientConn) debug(msg ...any) {
	wlog.WsDebug.Printf(cc.clientMsgFormat(msg...))
}
