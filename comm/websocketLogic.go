package comm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func wsConnect(ctx *gin.Context) {
	if ClientService == nil {
		ctx.Status(http.StatusServiceUnavailable)
		return
	}

	ctx.Status(http.StatusSwitchingProtocols)
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.ErrTrace(err)
		return
	}

	err = conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	if err != nil {
		log.ErrTrace(err)
		return
	}

	_, buf, err := conn.ReadMessage()
	if err != nil {
		_ = conn.Close()
		return
	}

	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		log.ErrTrace(err)
		return
	}

	var auth WsAuthorize
	err = json.Unmarshal(buf, &auth)
	if err != nil {
		log.ErrTrace(err)
		// ctx.Status(comm.StatusBadRequest)
		return
	}
	user, instance, err := WebsocketAuth(ctx, []string{auth.Auth})
	if err != nil {
		log.ShowErr(err)
		return
	}

	if user != nil {
		client := ClientService.ClientConnect(conn, user)
		go wsMain(client)
	} else if instance != nil {
		client := ClientService.RemoteConnect(conn, instance)
		go wsMain(client)
	} else {
		log.Error.Println("Hat trick nil on WebsocketAuth", auth.Auth)
		return
	}
}

func wsMain(c *WsClient) {
	defer ClientService.ClientDisconnect(c)
	var switchboard func([]byte, *WsClient)

	if c.GetUser() != nil {
		switchboard = wsWebClientSwitchboard
	} else {
		switchboard = wsInstanceClientSwitchboard
	}

	for {
		_, buf, err := c.ReadOne()
		if err != nil {
			break
		}
		go switchboard(buf, c)
	}
}

func wsWebClientSwitchboard(msgBuf []byte, c *WsClient) {
	defer wsRecover(c)

	var msg WsRequestInfo
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
	case FolderSubscribe:
		{
			err = json.Unmarshal([]byte(msg.Content), &subInfo)
			if err != nil {
				log.ErrTrace(err)
				c.Error(errors.New("failed to parse subscribe request"))
			}

			folderSub := subInfo.(*folderSubscribeMeta)
			var share models.Share
			if folderSub.ShareId != "" {
				share = ShareService.Get(folderSub.ShareId)
				if share == nil {
					c.Error(errors.New("share not found"))
					return
				}
			}

			complete, result, err := ClientService.Subscribe(c, subInfo.GetKey(), FolderSubscribe, share)
			if err != nil {
				c.Error(err)
				return
			}

			if complete {
				Caster.PushTaskUpdate(
					TaskService.GetTask(task.TaskId(subInfo.GetKey())), TaskCompleteEvent,
					result,
				)
			}
		}
	case TaskSubscribe:
		key := subInfo.GetKey()
		if key == "" {
			return
		}

		if strings.HasPrefix(string(key), "TID#") {
			key = key[4:]
			complete, result, err := ClientService.Subscribe(c, key, TaskSubscribe, nil)
			if err != nil {
				c.Error(err)
				return
			}

			if complete {
				Caster.PushTaskUpdate(
					TaskService.GetTask(task.TaskId(key)), TaskCompleteEvent,
					result,
				)
			}
		} else if strings.HasPrefix(string(key), "TT#") {
			key = key[3:]

			ClientService.Subscribe(c, key, TaskTypeSubscribe, nil)
		}

	case Unsubscribe:
		key := subInfo.GetKey()
		if key == "" {
			return
		}

		if strings.HasPrefix(string(key), "TID#") {
			key = key[4:]
		} else if strings.HasPrefix(string(key), "TT#") {
			key = key[3:]
		}

		err = ClientService.Unsubscribe(c, key)
		if err != nil {
			c.Error(err)
			return
		}

	case ScanDirectory:
		{
			if InstanceService.GetLocal().ServerRole() == models.BackupServer {
				return
			}

			folder, err := FileService.GetFileSafe(fileTree.FileId(subInfo.GetKey()), c.GetUser(), nil)
			if err != nil {
				c.Error(errors.New("could not find directory to scan"))
				return
			}

			newCaster := NewBufferedCaster(ClientService)
			meta := models.ScanMeta{
				File:         folder,
				FileService:  FileService,
				MediaService: MediaService,
				TaskService:  TaskService,
				Caster:       newCaster,
				TaskSubber:   ClientService,
			}
			t, err := TaskService.DispatchJob(models.ScanDirectoryTask, meta, nil)
			if err != nil {
				c.Error(err)
				return
			}
			t.SetCleanup(
				func() {
					newCaster.Close()
				},
			)

			_, _, err = ClientService.Subscribe(c, SubId(t.TaskId()), TaskSubscribe, nil)
			if err != nil {
				c.Error(err)
				return
			}
		}

	case CancelTask:
		{
			tpId := subInfo.GetKey()
			taskPool := TaskService.GetTaskPool(task.TaskId(tpId))
			if taskPool == nil {
				c.Error(errors.New("could not find task pool to cancel"))
				return
			}

			taskPool.Cancel()
			c.PushTaskUpdate(taskPool.CreatedInTask(), TaskCanceledEvent, nil)
		}

	default:
		{
			c.Error(fmt.Errorf("unknown websocket request type: %s", string(msg.Action)))
		}
	}
}

func wsInstanceClientSwitchboard(msgBuf []byte, c *WsClient) {
	defer wsRecover(c)

	var msg WsResponseInfo
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		c.Error(err)
		return
	}

	Caster.Relay(msg)
}

func wsRecover(c Client) {
	internal.RecoverPanic(fmt.Sprintf("[%s] websocket panic", c.GetUser().GetUsername()))
}
