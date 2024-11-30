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

func wsConnect(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	if pack.ClientService == nil || pack.AccessService == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	var u *models.User
	var err error
	var server *models.Instance
	getServer := r.URL.Query().Get("server") == "true"
	if getServer {
		server = getInstanceFromCtx(r)
		if server == nil {
			log.Error.Println("Got server websocket query but no server in context")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	} else {
		u, err = getUserFromCtx(r)
		if SafeErrorAndExit(err, w) {
			return
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if SafeErrorAndExit(err, w) {
		return
	}

	var client *models.WsClient
	if server != nil {
		client = pack.ClientService.RemoteConnect(conn, server)
	} else if u != nil {
		client = pack.ClientService.ClientConnect(conn, u)
	} else {
		// this should not happen
		log.Error.Println("Did not get valid websocket client")
		return
	}

	go wsMain(client, pack)
}

func wsMain(c *models.WsClient, pack *models.ServicePack) {
	defer pack.ClientService.ClientDisconnect(c)
	defer wsRecover(c)
	var switchboard func([]byte, *models.WsClient, *models.ServicePack)

	if c.GetUser() != nil {
		switchboard = wsWebClientSwitchboard
		onWebConnect(c, pack)
	} else {
		switchboard = wsServerClientSwitchboard
		if pack.Loaded.Load() {
			c.PushWeblensEvent(models.WeblensLoadedEvent, models.WsC{"role": pack.InstanceService.GetLocal().GetRole()})
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

	if pack.InstanceService.GetLocal().GetRole() == models.InitServerRole {
		c.Error(werror.ErrServerNotInitialized)
		return
	}

	var msg models.WsRequestInfo
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		log.Debug.Println(string(msgBuf))
		c.Error(werror.WithStack(err))
		return
	}

	log.Debug.Func(func(l log.Logger) { l.Printf("Got wsmsg from [%s]: %v", c.GetUser().GetUsername(), msg) })

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
					pack.TaskService.GetTask(subInfo.GetKey()), models.TaskCompleteEvent, result)
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
			if pack.InstanceService.GetLocal().GetRole() == models.BackupServerRole {
				return
			}

			folder, err := pack.FileService.GetFileSafe(subInfo.GetKey(), c.GetUser(), nil)
			if err != nil {
				c.Error(errors.New("could not find directory to scan"))
				return
			}

			meta := models.ScanMeta{
				File:         folder,
				FileService:  pack.FileService,
				MediaService: pack.MediaService,
				TaskService:  pack.TaskService,
				TaskSubber:   pack.ClientService,
			}

			var taskName string
			if folder.IsDir() {
				taskName = models.ScanDirectoryTask
			} else {
				taskName = models.ScanFileTask
				meta.Caster = pack.Caster
			}

			t, err := pack.TaskService.DispatchJob(taskName, meta, nil)
			if err != nil {
				c.Error(err)
				return
			}

			_, _, err = pack.ClientService.Subscribe(c, t.TaskId(), models.TaskSubscribe, time.Now(), nil)
			if err != nil {
				c.Error(err)
				return
			}
		}

	case models.CancelTask:
		{
			taskId := subInfo.GetKey()
			task := pack.TaskService.GetTask(taskId)
			if task == nil {
				c.Error(werror.Errorf("could not find task T[%s] to cancel", taskId))
				return
			}

			task.Cancel()
			c.PushTaskUpdate(task, models.TaskCanceledEvent, nil)
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

	if msg.SentTime == 0 {
		err := werror.Errorf("invalid sent time on relay message %s", msg.EventTag)
		c.Error(err)
		return
	}

	sentTime := time.UnixMilli(msg.SentTime)
	relaySourceId := c.GetRemote().ServerId()

	switch msg.EventTag {
	case models.ServerGoingDownEvent:
		c.Disconnect()
		return
	case models.BackupCompleteEvent:
		// Log the backup time, but don't return so the
		// message can be relayed to the web client
		err := pack.InstanceService.SetLastBackup(relaySourceId, sentTime)
		if err != nil {
			c.Error(err)
			return
		}

		if pack.InstanceService.GetLocal().IsCore() {
			// Also update the local core server's last backup time
			err = pack.InstanceService.SetLastBackup(pack.InstanceService.GetLocal().ServerId(), sentTime)
			if err != nil {
				c.Error(err)
				return
			}
		}
	case models.RemoteConnectionChangedEvent:
		return
	}

	msg.RelaySource = relaySourceId
	pack.Caster.Relay(msg)
}

func onWebConnect(c models.Client, pack *models.ServicePack) {
	if !pack.Loaded.Load() {
		c.PushWeblensEvent(models.StartupProgressEvent, models.WsC{"waitingOn": pack.GetStartupTasks()})
		return
	} else {
		c.PushWeblensEvent(models.WeblensLoadedEvent, models.WsC{"role": pack.InstanceService.GetLocal().GetRole()})
	}

	if pack.InstanceService.GetLocal().GetRole() == models.BackupServerRole {
		for _, backupTask := range pack.TaskService.GetTasksByJobName(models.BackupTask) {
			r := backupTask.GetResults()
			if len(r) == 0 {
				continue
			}

			c.PushTaskUpdate(backupTask, models.BackupProgressEvent, r)
		}
	}
}

func wsRecover(c models.Client) {
	internal.RecoverPanic(fmt.Sprintf("[%s] websocket panic", c.GetUser().GetUsername()))
}
