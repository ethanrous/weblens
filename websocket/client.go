package websocket

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"slices"
	"sync"

	weblens "github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/fileTree"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/gorilla/websocket"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
)

type ClientId string

type WsClient struct {
	Active        bool
	connId ClientId
	conn          *websocket.Conn
	mu            sync.Mutex
	subscriptions []types.Subscription
	user   *weblens.User
	remote *weblens.WeblensInstance
}

func (wsc *WsClient) IsOpen() bool {
	return wsc.Active
}

func (wsc *WsClient) GetClientId() ClientId {
	return wsc.connId
}

func (wsc *WsClient) ClientType() types.ClientType {
	return types.WebClient
}

func (wsc *WsClient) GetShortId() ClientId {
	if wsc.connId == "" {
		return ""
	}
	return wsc.connId[28:]
}

func (wsc *WsClient) GetUser() *weblens.User {
	return wsc.user
}

func (wsc *WsClient) GetRemote() *weblens.WeblensInstance {
	return wsc.remote
}

func (wsc *WsClient) ReadOne() (int, []byte, error) {
	return wsc.conn.ReadMessage()
}

func (wsc *WsClient) Disconnect() {
	wsc.Active = false
	types.SERV.ClientManager.ClientDisconnect(wsc)

	wsc.mu.Lock()
	err := wsc.conn.Close()
	if err != nil {
		wlog.ShowErr(err)
		return
	}
	wsc.mu.Unlock()
	wsc.log("Disconnected")
}

// Subscribe links a websocket connection to a "key" that can be broadcast to later if
// relevant updates should be communicated
//
// Returns "true" and the results at meta.LookingFor if the task is completed, false and nil otherwise.
// Subscriptions to types that represent ongoing events like "folder" never return truthy completed
func (wsc *WsClient) Subscribe(
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
				wsc.Error(err)
				return
			}
			fileId := fileTree.FileId(key)
			var folder *fileTree.WeblensFile
			if fileId == "external" {
				folder = types.SERV.FileTree.Get("EXTERNAL")
			} else {
				folder = types.SERV.FileTree.Get(fileId)
			}

			acc.SetRequestMode(dataStore.FileSubscribeRequest)

			if folder == nil {
				err := werror.WErrMsg(fmt.Sprint("failed to find folder to subscribe to: ", key))
				wsc.Error(err)
				return
			} else if !acc.CanAccessFile(folder) {
				err := werror.WErrMsg(fmt.Sprint("failed to find folder to subscribe to: ", key))
				wsc.Error(err)

				// don't tell the clientConn they don't have access instead of *actually* not found
				wsc.err("User does not have access to", key)
				return
			}

			sub = types.Subscription{Type: types.FolderSubscribe, Key: key}
			wsc.PushFileUpdate(folder)

			// Subscribe to task on this folder
			if t := folder.GetTask(); t != nil {
				wsc.Subscribe(types.SubId(t.TaskId()), types.TaskSubscribe, acc)
			}
		}
	case types.TaskSubscribe:
		{
			task := types.SERV.TaskDispatcher.GetWorkerPool().GetTask(types.TaskId(key))
			if task == nil {
				err := fmt.Errorf("could not find task with ID %s", key)
				wsc.Error(err)
				return
			}

			complete, _ = task.Status()
			results = task.GetResults()

			wsc.mu.Lock()
			if complete || slices.IndexFunc(
				wsc.subscriptions, func(s types.Subscription) bool { return s.Key == key },
			) != -1 {
				wsc.mu.Unlock()
				return
			}
			wsc.mu.Unlock()

			sub = types.Subscription{Type: types.TaskSubscribe, Key: key}

			wsc.PushTaskUpdate(task, dataProcess.TaskCreatedEvent, task.GetMeta().FormatToResult())
		}
	case types.PoolSubscribe:
		{
			pool := types.SERV.WorkerPool.GetTaskPool(types.TaskId(key))
			if pool == nil {
				wsc.Error(werror.NewWeblensError(fmt.Sprintf("Could not find pool with id %s", key)))
				return
			} else if pool.IsGlobal() {
				wsc.Error(werror.NewWeblensError("Trying to subscribe to global pool"))
				return
			}

			wlog.Debug.Printf("%s subscribed to pool [%s]", wsc.user.GetUsername(), pool.ID())
			sub = types.Subscription{Type: types.TaskSubscribe, Key: key}

			wsc.PushPoolUpdate(
				pool, dataProcess.PoolCreatedEvent, types.TaskResult{
					"createdBy": pool.CreatedInTask().
						TaskId(),
				},
			)
		}
	case types.TaskTypeSubscribe:
		{
			wlog.Debug.Printf("%s subscribed to task type [%s]", wsc.user.GetUsername(), key)
			sub = types.Subscription{Type: types.TaskTypeSubscribe, Key: key}
		}
	default:
		{
			err := fmt.Errorf("unknown subscription type %s", action)
			wlog.ErrTrace(err)
			wsc.Error(err)
			return
		}
	}
	wsc.debug("Subscribed to", action, key)

	wsc.mu.Lock()
	wsc.subscriptions = append(wsc.subscriptions, sub)
	wsc.mu.Unlock()
	types.SERV.ClientManager.AddSubscription(sub, wsc)

	return
}

func (wsc *WsClient) Unsubscribe(key types.SubId) {
	wsc.mu.Lock()
	subIndex := slices.IndexFunc(wsc.subscriptions, func(s types.Subscription) bool { return s.Key == key })
	if subIndex == -1 {
		wsc.mu.Unlock()
		return
	}
	var subToRemove types.Subscription
	wsc.subscriptions, subToRemove = internal.Yoink(wsc.subscriptions, subIndex)
	wsc.mu.Unlock()

	types.SERV.ClientManager.RemoveSubscription(subToRemove, wsc, false)

	wsc.debug("Unsubscribed from", subToRemove.Type, subToRemove.Key)
}

func (wsc *WsClient) Error(err error) {
	var weblensError werror.WErr
	ok := errors.As(err, &weblensError)

	var msg types.WsResponseInfo
	switch ok {
	case true:
		wsc.err(err)
		msg = types.WsResponseInfo{EventTag: "error", Error: err.Error()}
	case false:
		wsc.errTrace(err)
		msg = types.WsResponseInfo{EventTag: "error", Error: "Masked unexpected server error"}
	}

	wsc.Send(msg)
}

func (wsc *WsClient) PushWeblensEvent(eventTag string) {
	msg := types.WsResponseInfo{
		EventTag:      eventTag,
		SubscribeKey:  types.SubId("WEBLENS"),
		BroadcastType: types.ServerEvent,
	}

	wsc.Send(msg)
}

func (wsc *WsClient) PushFileUpdate(updatedFile *fileTree.WeblensFile) {
	acc := dataStore.NewAccessMeta(wsc.user).SetRequestMode(dataStore.WebsocketFileUpdate)
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

	wsc.Send(msg)
}

func (wsc *WsClient) PushTaskUpdate(task types.Task, event types.TaskEvent, result types.TaskResult) {
	msg := types.WsResponseInfo{
		EventTag:      string(event),
		SubscribeKey:  types.SubId(task.TaskId()),
		Content:       types.WsC(result),
		TaskType:      task.TaskType(),
		BroadcastType: types.TaskSubscribe,
	}

	wsc.Send(msg)
	// c.Send(string(event), types.SubId(taskId), []types.WsC{types.WsC(result)})
}

func (wsc *WsClient) PushPoolUpdate(pool types.TaskPool, event types.TaskEvent, result types.TaskResult) {
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

	wsc.Send(msg)
	// c.Send(string(event), types.SubId(taskId), []types.WsC{types.WsC(result)})
}

func (wsc *WsClient) GetSubscriptions() []types.Subscription {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()
	return wsc.subscriptions[0:len(wsc.subscriptions)]
}

func (wsc *WsClient) Send(msg types.WsResponseInfo) {
	if wsc != nil {
		wsc.mu.Lock()
		defer wsc.mu.Unlock()
		err := wsc.conn.WriteJSON(msg)
		if err != nil {
			wsc.errTrace(err)
		}
	}
}

func (wsc *WsClient) Raw(msg any) error {
	return wsc.conn.WriteJSON(msg)
}

func (wsc *WsClient) clientMsgFormat(msg ...any) string {
	var clientName string
	if wsc.GetUser() != nil {
		clientName = string(wsc.GetUser().GetUsername())
	} else {
		clientName = wsc.GetRemote().GetName()
	}
	return fmt.Sprintf("| %s (%s) | %s", wsc.GetShortId(), clientName, fmt.Sprintln(msg...))
}

func (wsc *WsClient) log(msg ...any) {
	wlog.WsInfo.Printf(wsc.clientMsgFormat(msg...))
}

func (wsc *WsClient) err(msg ...any) {
	_, file, line, _ := runtime.Caller(2)
	msg = []any{any(fmt.Sprintf("%s:%d:", file, line)), msg}
	wlog.WsError.Printf(wsc.clientMsgFormat(msg...))
}

func (wsc *WsClient) errTrace(msg ...any) {
	_, file, line, _ := runtime.Caller(2)
	msg = []any{any(fmt.Sprintf("%s:%d:", file, line)), msg}
	wlog.WsError.Printf(wsc.clientMsgFormat(msg...), string(debug.Stack()))
}

func (wsc *WsClient) debug(msg ...any) {
	wlog.WsDebug.Printf(wsc.clientMsgFormat(msg...))
}
