package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func wsConnect(ctx *gin.Context) {
	ctx.Status(http.StatusSwitchingProtocols)
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		util.ErrTrace(err)
		return
	}

	_, buf, err := conn.ReadMessage()
	if err != nil {
		return
	}

	var auth wsAuthorize
	err = json.Unmarshal(buf, &auth)
	if err != nil {
		util.ErrTrace(err)
		// ctx.Status(http.StatusBadRequest)
		return
	}
	user, err := WebsocketAuth(ctx, []string{auth.Auth})
	if err != nil {
		util.ShowErr(err)
		return
	}

	client := types.SERV.ClientManager.ClientConnect(conn, user)
	go wsMain(client)
}

func wsMain(c types.Client) {
	defer c.Disconnect()

	for {
		_, buf, err := c.(*client).conn.ReadMessage()
		if err != nil {
			break
		}
		go wsReqSwitchboard(buf, c)
	}
}

func wsReqSwitchboard(msgBuf []byte, c types.Client) {
	defer wsRecover(c)
	// defer util.RecoverPanic("[WS] client %d panicked: %v", client.GetClientId())

	var msg wsRequest
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
				util.ErrTrace(err)
				c.Error(errors.New("failed to parse subscribe request"))
			}

			folderSub := subInfo.(*folderSubscribeMeta)
			acc := dataStore.NewAccessMeta(c.GetUser(), types.SERV.FileTree)
			if folderSub.ShareId != "" {
				sh := types.SERV.ShareService.Get(folderSub.ShareId)
				if sh == nil {
					c.Error(types.NewWeblensError("share not found"))
					return
				}

				err = acc.AddShare(sh)
				if err != nil {
					util.ErrTrace(err)
					c.Error(errors.New("failed to add share"))
					return
				}
			}

			complete, result := c.Subscribe(subInfo.GetKey(), types.FolderSubscribe, acc, types.SERV.FileTree)
			if complete {
				types.SERV.Caster.PushTaskUpdate(
					types.SERV.WorkerPool.GetTask(types.TaskId(subInfo.GetKey())), dataProcess.TaskCompleteEvent,
					types.TaskResult{"takeoutId": result["takeoutId"]},
				)
			}
		}
	case types.TaskSubscribe:
		complete, result := c.Subscribe(subInfo.GetKey(), types.TaskSubscribe, nil, types.SERV.FileTree)
		if complete {
			types.SERV.Caster.PushTaskUpdate(
				types.SERV.WorkerPool.GetTask(types.TaskId(subInfo.GetKey())), dataProcess.TaskCompleteEvent,
				types.TaskResult{"takeoutId": result["takeoutId"]},
			)
		}
	case types.Unsubscribe:
		c.Unsubscribe(subInfo.GetKey())

	case types.ScanDirectory:
		{
			folder := types.SERV.FileTree.Get(types.FileId(subInfo.GetKey()))
			if folder == nil {
				c.Error(types.NewWeblensError("could not find directory to scan"))
				return
			}

			c.(*client).debug("Got scan directory for", folder.GetAbsPath())

			newCaster := NewBufferedCaster()
			t := types.SERV.TaskDispatcher.ScanDirectory(folder, newCaster)
			t.SetCleanup(
				func() {
					newCaster.Close()
				},
			)
			acc := dataStore.NewAccessMeta(c.GetUser(), types.SERV.FileTree)
			c.Subscribe(types.SubId(t.TaskId()), types.TaskSubscribe, acc, types.SERV.FileTree)
		}

	case types.CancelTask:
		{
			tpId := subInfo.GetKey()
			taskPool := types.SERV.WorkerPool.GetTaskPool(types.TaskId(tpId))
			if taskPool == nil {
				c.Error(types.NewWeblensError("could not find task pool to cancel"))
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

func wsRecover(c types.Client) {
	err := recover()
	if err != nil {
		c.Error(fmt.Errorf("%v", err))
	}
}
