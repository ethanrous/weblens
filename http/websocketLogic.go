package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/task"
	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
)

type WsAuthorize struct {
	Auth string `json:"auth"`
}

var upgrader = gorilla.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsConnect(ctx *gin.Context) {
	pack := getServices(ctx)
	if pack.ClientService == nil || pack.AccessService == nil {
		ctx.Status(http.StatusServiceUnavailable)
		return
	}

	ctx.Status(http.StatusSwitchingProtocols)
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.ErrTrace(err)
		return
	}

	usr := getUserFromCtx(ctx)
	server := getInstanceFromCtx(ctx)

	var client *models.WsClient
	if usr != nil {
		client = pack.ClientService.ClientConnect(conn, usr)
	} else if server != nil {
		client = pack.ClientService.RemoteConnect(conn, server)
	} else {
		// this should not happen
		log.Error.Println("Did not get valid websocket client")
		return
	}

	go wsMain(client, pack)
}

func wsMain(c *models.WsClient, pack *models.ServicePack) {
	defer pack.ClientService.ClientDisconnect(c)
	var switchboard func([]byte, *models.WsClient, *models.ServicePack)

	if c.GetUser() != nil {
		switchboard = wsWebClientSwitchboard
		if !pack.Loaded.Load() {
			c.PushWeblensEvent(models.StartupProgressEvent, models.WsC{"waitingOn": pack.GetStartupTasks()})
		} else {
			c.PushWeblensEvent("weblens_loaded", models.WsC{"role": pack.InstanceService.GetLocal().GetRole()})
		}
	} else {
		switchboard = wsServerClientSwitchboard
		if pack.Loaded.Load() {
			c.PushWeblensEvent("weblens_loaded", models.WsC{"role": pack.InstanceService.GetLocal().GetRole()})
		}
	}

	for {
		_, buf, err := c.ReadOne()
		if err != nil {
			break
		}
		go switchboard(buf, c, pack)
	}
}

func wsWebClientSwitchboard(msgBuf []byte, c *models.WsClient, pack *models.ServicePack) {
	defer wsRecover(c)

	if pack.InstanceService.GetLocal().GetRole() == models.InitServer {
		c.Error(werror.ErrServerNotInitialized)
		return
	}

	var msg models.WsRequestInfo
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		c.Error(werror.WithStack(err))
		return
	}

	log.Trace.Printf("Got wsmsg from [%s]: %v", c.GetUser().GetUsername(), msg)

	if msg.Action == models.ReportError {
		log.ErrorCatcher.Printf("Web client caught unexpected error\n%s\n\n", msg.Content)
		return
	}

	subInfo, err := newActionBody(msg)
	if err != nil {
		c.Error(err)
		return
	}

	switch subInfo.Action() {
	case models.FolderSubscribe:
		{
			if pack.FileService == nil {
				c.Error(werror.Errorf("file service not ready"))
				return
			}

			err = json.Unmarshal([]byte(msg.Content), &subInfo)
			if err != nil {
				log.ErrTrace(err)
				c.Error(errors.New("failed to parse subscribe request"))
			}

			folderSub := subInfo.(*folderSubscribeMeta)
			var share models.Share
			if folderSub.ShareId != "" {
				share = pack.ShareService.Get(folderSub.ShareId)
				if share == nil {
					c.Error(errors.New("share not found"))
					return
				}
			}

			complete, result, err := pack.ClientService.Subscribe(
				c, subInfo.GetKey(), models.FolderSubscribe, time.UnixMilli(msg.SentAt), share,
			)
			if err != nil {
				c.Error(err)
				return
			}

			if complete {
				pack.Caster.PushTaskUpdate(
					pack.TaskService.GetTask(subInfo.GetKey()), models.TaskCompleteEvent,
					result,
				)
			}
		}
	case models.TaskSubscribe:
		key := subInfo.GetKey()
		if key == "" {
			return
		}

		if strings.HasPrefix(key, "TID#") {
			key = key[4:]
			complete, result, err := pack.ClientService.Subscribe(c, key, models.TaskSubscribe, time.Now(), nil)
			if err != nil {
				c.Error(err)
				return
			}

			if complete {
				pack.Caster.PushTaskUpdate(
					pack.TaskService.GetTask(key), models.TaskCompleteEvent,
					result,
				)
			}
		} else if strings.HasPrefix(key, "TT#") {
			key = key[3:]

			_, _, err := pack.ClientService.Subscribe(c, key, models.TaskTypeSubscribe, time.Now(), nil)
			if err != nil {
				c.Error(err)
			}
		}

	case models.Unsubscribe:
		key := subInfo.GetKey()
		if key == "" {
			return
		}

		if strings.HasPrefix(key, "TID#") {
			key = key[4:]
		} else if strings.HasPrefix(key, "TT#") {
			key = key[3:]
		}

		err = pack.ClientService.Unsubscribe(c, key, time.UnixMilli(msg.SentAt))
		if err != nil {
			c.Error(err)
			return
		}

	case models.ScanDirectory:
		{
			if pack.InstanceService.GetLocal().GetRole() == models.BackupServer {
				return
			}

			folder, err := pack.FileService.GetFileSafe(subInfo.GetKey(), c.GetUser(), nil)
			if err != nil {
				c.Error(errors.New("could not find directory to scan"))
				return
			}

			newCaster := models.NewSimpleCaster(pack.ClientService)
			meta := models.ScanMeta{
				File:         folder,
				FileService:  pack.FileService,
				MediaService: pack.MediaService,
				TaskService:  pack.TaskService,
				Caster:       newCaster,
				TaskSubber:   pack.ClientService,
			}

			var taskName string
			if folder.IsDir() {
				taskName = models.ScanDirectoryTask
			} else {
				taskName = models.ScanFileTask
			}

			t, err := pack.TaskService.DispatchJob(taskName, meta, nil)
			if err != nil {
				c.Error(err)
				return
			}
			t.SetCleanup(
				func(t *task.Task) {
					newCaster.Close()
				},
			)

			_, _, err = pack.ClientService.Subscribe(c, t.TaskId(), models.TaskSubscribe, time.Now(), nil)
			if err != nil {
				c.Error(err)
				return
			}
		}

	case models.CancelTask:
		{
			tpId := subInfo.GetKey()
			taskPool := pack.TaskService.GetTaskPool(tpId)
			if taskPool == nil {
				c.Error(errors.New("could not find task pool to cancel"))
				return
			}

			taskPool.Cancel()
			c.PushTaskUpdate(taskPool.CreatedInTask(), models.TaskCanceledEvent, nil)
		}

	default:
		{
			c.Error(fmt.Errorf("unknown websocket request type: %s", string(msg.Action)))
		}
	}
}

func wsServerClientSwitchboard(msgBuf []byte, c *models.WsClient, pack *models.ServicePack) {
	defer wsRecover(c)

	var msg models.WsResponseInfo
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		c.Error(err)
		return
	}

	if msg.EventTag == models.ServerGoingDownEvent {
		c.Disconnect()
		return
	}

	pack.Caster.Relay(msg)
}

func wsRecover(c models.Client) {
	internal.RecoverPanic(fmt.Sprintf("[%s] websocket panic", c.GetUser().GetUsername()))
}
