package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
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

	var usr *models.User
	var instance *models.Instance
	if len(auth.Auth) != 0 && auth.Auth[0] == 'X' {
		usr, instance, err = ParseApiKeyLogin(auth.Auth, pack)
		if err != nil {
			log.ShowErr(err)
			return
		}
	} else {
		usr, err = ParseUserLogin(auth.Auth, pack.AccessService)
		if err != nil {
			log.ShowErr(err)
			return
		}
	}

	var client *models.WsClient
	if usr != nil {
		client = pack.ClientService.ClientConnect(conn, usr)
	} else if instance != nil {
		client = pack.ClientService.RemoteConnect(conn, instance)
	} else {
		// this should not happen
		log.Error.Println("Hat trick nil on WebsocketAuth", auth.Auth)
		return
	}

	go wsMain(client, pack)
}

func wsMain(c *models.WsClient, pack *models.ServicePack) {
	defer pack.ClientService.ClientDisconnect(c)
	var switchboard func([]byte, *models.WsClient, *models.ServicePack)

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
		go switchboard(buf, c, pack)
	}
}

func wsWebClientSwitchboard(msgBuf []byte, c *models.WsClient, pack *models.ServicePack) {
	defer wsRecover(c)

	var msg models.WsRequestInfo
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		c.Error(err)
		return
	}

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
			if pack.InstanceService.GetLocal().ServerRole() == models.BackupServer {
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

func wsInstanceClientSwitchboard(msgBuf []byte, c *models.WsClient, pack *models.ServicePack) {
	defer wsRecover(c)

	var msg models.WsResponseInfo
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		c.Error(err)
		return
	}

	pack.Caster.Relay(msg)
}

func wsRecover(c models.Client) {
	internal.RecoverPanic(fmt.Sprintf("[%s] websocket panic", c.GetUser().GetUsername()))
}
