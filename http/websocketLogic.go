package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	error2 "github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/websocket"
	"github.com/gin-gonic/gin"
)

func wsConnect(ctx *gin.Context) {
	if types.SERV.GetClientServiceSafely() == nil {
		ctx.Status(http.StatusServiceUnavailable)
		return
	}

	ctx.Status(http.StatusSwitchingProtocols)
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		wlog.ErrTrace(err)
		return
	}

	err = conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	if err != nil {
		wlog.ErrTrace(err)
		return
	}

	_, buf, err := conn.ReadMessage()
	if err != nil {
		_ = conn.Close()
		return
	}

	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		wlog.ErrTrace(err)
		return
	}

	var auth wsAuthorize
	err = json.Unmarshal(buf, &auth)
	if err != nil {
		wlog.ErrTrace(err)
		// ctx.Status(http.StatusBadRequest)
		return
	}
	user, instance, err := WebsocketAuth(ctx, []string{auth.Auth})
	if err != nil {
		wlog.ShowErr(err)
		return
	}

	if user != nil {
		client := types.SERV.ClientManager.ClientConnect(conn, user)
		go wsMain(client)
	} else if instance != nil {
		client := types.SERV.ClientManager.RemoteConnect(conn, instance)
		go wsMain(client)
	} else {
		wlog.Error.Println("Hat trick nil on WebsocketAuth", auth.Auth)
		return
	}
}

func wsMain(c types.Client) {
	defer c.Disconnect()
	var switchboard func([]byte, types.Client)

	if c.GetUser() != nil {
		switchboard = wsWebClientSwitchboard
	} else {
		switchboard = wsInstanceClientSwitchboard
	}

	for {
		_, buf, err := c.(*websocket.clientConn).conn.ReadMessage()
		if err != nil {
			break
		}
		go switchboard(buf, c)
	}
}

func wsWebClientSwitchboard(msgBuf []byte, c types.Client) {
	defer wsRecover(c)

	var msg types.WsRequestInfo
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		c.Error(err)
		return
	}

	subInfo, err := newActionBody(msg)
	if err != nil {
		c.Error(err)
		return
	}

	switch subInfo.Action() {
	case types.FolderSubscribe:
		{
			err = json.Unmarshal([]byte(msg.Content), &subInfo)
			if err != nil {
				wlog.ErrTrace(err)
				c.Error(errors.New("failed to parse subscribe request"))
			}

			folderSub := subInfo.(*folderSubscribeMeta)
			acc := dataStore.NewAccessMeta(c.GetUser())
			if folderSub.ShareId != "" {
				sh := types.SERV.ShareService.Get(folderSub.ShareId)
				if sh == nil {
					c.Error(error2.WErrMsg("share not found"))
					return
				}

				err = acc.AddShare(sh)
				if err != nil {
					wlog.ErrTrace(err)
					c.Error(error2.WErrMsg("failed to add share"))
					return
				}
			}

			complete, result := c.Subscribe(subInfo.GetKey(), types.FolderSubscribe, acc)
			if complete {
				types.SERV.Caster.PushTaskUpdate(
					types.SERV.WorkerPool.GetTask(types.TaskId(subInfo.GetKey())), dataProcess.TaskCompleteEvent,
					result,
				)
			}
		}
	case types.TaskSubscribe:
		key := subInfo.GetKey()
		if key == "" {
			return
		}

		if strings.HasPrefix(string(key), "TID#") {
			key = key[4:]
			complete, result := c.Subscribe(key, types.TaskSubscribe, nil)
			if complete {
				types.SERV.Caster.PushTaskUpdate(
					types.SERV.WorkerPool.GetTask(types.TaskId(key)), dataProcess.TaskCompleteEvent,
					result,
				)
			}
		} else if strings.HasPrefix(string(key), "TT#") {
			key = key[3:]

			c.Subscribe(key, types.TaskTypeSubscribe, nil)

			// pool := types.SERV.WorkerPool.GetTaskPoolByTaskType(types.TaskType(key))
			// complete, result := c.Subscribe(types.SubId(pool.ID()), types.PoolSubscribe, nil)
			// if complete {
			// 	types.SERV.Caster.PushTaskUpdate(
			// 		types.SERV.WorkerPool.GetTaskPool(types.TaskId(key)).CreatedInTask(),
			// 		dataProcess.PoolCompleteEvent,
			// 		result,
			// 	)
			// }
		}

	case types.Unsubscribe:
		key := subInfo.GetKey()
		if key == "" {
			return
		}
		
		if strings.HasPrefix(string(key), "TID#") {
			key = key[4:]
		} else if strings.HasPrefix(string(key), "TT#") {
			key = key[3:]
		}

		c.Unsubscribe(key)

	case types.ScanDirectory:
		{
			if InstanceService.GetLocal().ServerRole() == BackupServer {
				return
			}

			folder := types.SERV.FileTree.Get(types.FileId(subInfo.GetKey()))
			if folder == nil {
				c.Error(error2.NewWeblensError("could not find directory to scan"))
				return
			}

			c.(*websocket.clientConn).debug("Got scan directory for", folder.GetAbsPath())

			newCaster := websocket.NewBufferedCaster()
			t := types.SERV.TaskDispatcher.ScanDirectory(folder, newCaster)
			t.SetCleanup(
				func() {
					newCaster.Close()
				},
			)
			acc := dataStore.NewAccessMeta(c.GetUser())
			c.Subscribe(types.SubId(t.TaskId()), types.TaskSubscribe, acc)
		}

	case types.CancelTask:
		{
			tpId := subInfo.GetKey()
			taskPool := types.SERV.WorkerPool.GetTaskPool(types.TaskId(tpId))
			if taskPool == nil {
				c.Error(error2.NewWeblensError("could not find task pool to cancel"))
				return
			}

			taskPool.Cancel()
			c.PushTaskUpdate(taskPool.CreatedInTask(), dataProcess.TaskCanceledEvent, nil)
		}

	default:
		{
			c.Error(fmt.Errorf("unknown websocket request type: %s", string(msg.Action)))
		}
	}
}

func wsInstanceClientSwitchboard(msgBuf []byte, c types.Client) {
	defer wsRecover(c)

	var msg types.WsResponseInfo
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		c.Error(err)
		return
	}

	types.SERV.Caster.(*websocket.unbufferedCaster).Relay(msg)
}

func wsRecover(c types.Client) {
	err := recover()
	if err != nil {
		c.Error(fmt.Errorf("%v", err))
	}
}
