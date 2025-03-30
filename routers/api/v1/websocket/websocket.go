package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/models"
	client_model "github.com/ethanrous/weblens/models/client"
	share_model "github.com/ethanrous/weblens/models/share"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/services/context"
	gorilla "github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const IsTowerQueryKey = "isTower"

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

func Connect(ctx context.RequestContext) {
	// if ctx.Query(IsTowerQueryKey) == "true" {
	// } else {
	// }

	// if getServer {
	// 	server = getInstanceFromCtx(r)
	// 	if server == nil {
	// 		SafeErrorAndExit(werror.WithStack(werror.ErrNoServerInContext), w)
	// 		return
	// 	}
	// } else {
	// 	u, err = getUserFromCtx(r, true)
	// 	if SafeErrorAndExit(err, w) {
	// 		return
	// 	}
	// }

	conn, err := upgrader.Upgrade(ctx.W, ctx.Req, nil)
	if err != nil {
		ctx.Logger.Error().Err(err).Msg("Failed to upgrade connection to websocket")
		return
	}

	ctx.ClientService.ClientConnect(conn, ctx.Requester)

	// var client *models.WsClient
	// if server != nil {
	// 	client = pack.ClientService.RemoteConnect(conn, server)
	// } else if u != nil {
	// 	client = pack.ClientService.ClientConnect(conn, u)
	// } else {
	// 	// this should not happen
	// 	pack.Log.Error().Msg("Did not get valid websocket client")
	// 	return
	// }

	go wsMain(ctx, client_model)
}

func wsMain(ctx context.RequestContext, c *client_model.WsClient) {
	defer ctx.ClientService.ClientDisconnect(c)
	defer wsRecover(c)
	var switchboard func(context.RequestContext, []byte, *client_model.WsClient) error

	if c.GetUser() != nil {
		switchboard = wsWebClientSwitchboard
		onWebConnect(c, pack)
	} else {
		switchboard = wsServerClientSwitchboard
		if pack.Loaded.Load() {
			pack.Log.Debug().Msgf("Server connected: %s -- local role is: %s", c.GetInstance().ServerId(), pack.InstanceService.GetLocal().GetRole())
			c.PushWeblensEvent(models.WeblensLoadedEvent, models.WsC{"role": pack.InstanceService.GetLocal().GetRole()})
		}
	}

	for {
		_, buf, err := c.ReadOne()
		if err != nil {
			break
		}
		go func() {
			defer wsRecover(c)
			err := switchboard(buf, c, pack)
			if err != nil {
				c.Error(err)
			}
		}()
	}
}

func wsWebClientSwitchboard(ctx context.RequestContext, msgBuf []byte, c *client_model.WsClient) error {
	// localTower, err := tower_model.GetLocal(ctx)
	// if err != nil {
	// 	return err
	// }

	// TODO: Verify that we want to ignore requests to the init server
	// if localTower.Role == tower_model.InitServerRole {
	// 	return werror.ErrServerNotInitialized
	// }

	var msg client_model.WsRequestInfo
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		return werror.WithStack(err)
	}

	c.Log().Trace().Func(func(e *zerolog.Event) { e.Interface("websocket_msg_content", msg).Msg("Got wsmsg from client") })

	if msg.Action == client_model.ReportError {
		c.Log().Error().Msgf("Web client reported error: %s", msg.Content)
		return nil
	}

	subInfo, err := newActionBody(msg)
	if err != nil {
		return err
	}

	switch subInfo.Action() {
	case client_model.FolderSubscribe:
		{
			if pack.FileService == nil {
				return werror.Errorf("file service not ready")
			}

			err = json.Unmarshal([]byte(msg.Content), &subInfo)
			if err != nil {
				return werror.WithStack(err)
			}

			share_model.GetShareById(ctx, subInfo.ShareId)

			share := subInfo.GetShare(pack.ShareService)
			complete, result, err := pack.ClientService.Subscribe(
				c, subInfo.GetKey(), client_model.FolderSubscribe, time.UnixMilli(msg.SentAt), share,
			)
			if err != nil {
				return err
			}

			if complete {
				pack.Caster.PushTaskUpdate(
					pack.TaskService.GetTask(subInfo.GetKey()), client_model.TaskCompleteEvent, result)
			}
		}
	case client_model.TaskSubscribe:
		key := subInfo.GetKey()
		if key == "" {
			return werror.WithStack(werror.ErrNoSubKey)
		}

		if strings.HasPrefix(key, "TID#") {
			key = key[4:]
			complete, result, err := pack.ClientService.Subscribe(c, key, client_model.TaskSubscribe, time.Now(), nil)
			if err != nil {
				return err
			}

			if complete {
				pack.Caster.PushTaskUpdate(
					pack.TaskService.GetTask(key), client_model.TaskCompleteEvent,
					result,
				)
			}
		} else if strings.HasPrefix(key, "TT#") {
			key = key[3:]

			_, _, err := pack.ClientService.Subscribe(c, key, client_model.TaskTypeSubscribe, time.Now(), nil)
			if err != nil {
				c.Error(err)
			}
		}

	case client_model.Unsubscribe:
		key := subInfo.GetKey()
		if key == "" {
			return werror.WithStack(werror.ErrNoSubKey)
		}

		if strings.HasPrefix(key, "TID#") {
			key = key[4:]
		} else if strings.HasPrefix(key, "TT#") {
			key = key[3:]
		}

		err = pack.ClientService.Unsubscribe(c, key, time.UnixMilli(msg.SentAt))
		if err != nil && !errors.Is(err, werror.ErrSubscriptionNotFound) {
			c.Error(err)
		} else if err != nil {
			c.Log().Warn().Msgf("Subscription [%s] not found in unsub task", key)
		}

	case client_model.ScanDirectory:
		{
			if pack.InstanceService.GetLocal().GetRole() == models.BackupServerRole {
				return werror.ErrServerIsBackup
			}

			share, err := share.GetShareById(ctx, subInfo.ShareId)

			folder, err := pack.FileService.GetFileSafe(subInfo.GetKey(), c.GetUser(), share)
			if err != nil {
				return werror.Errorf("could not find directory to scan")
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
				return err
			}

			_, _, err = pack.ClientService.Subscribe(c, t.TaskId(), client_model.TaskSubscribe, time.Now(), share)
			if err != nil {
				return err
			}
		}

	case client_model.CancelTask:
		{
			taskId := subInfo.GetKey()
			task := pack.TaskService.GetTask(taskId)
			if task == nil {
				return werror.Errorf("could not find task T[%s] to cancel", taskId)
			}

			task.Cancel()
			c.PushTaskUpdate(task, client_model.TaskCanceledEvent, nil)
		}

	default:
		{
			c.Error(fmt.Errorf("unknown websocket request type: %s", string(msg.Action)))
		}
	}

	return nil
}

func wsServerClientSwitchboard(ctx context.RequestContext, msgBuf []byte, c *client_model.WsClient) error {
	defer wsRecover(c)

	var msg client_model.WsResponseInfo
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		return err
	}

	if msg.SentTime == 0 {
		err := werror.Errorf("invalid sent time on relay message: [%s]", msg.EventTag)
		return err
	}

	sentTime := time.UnixMilli(msg.SentTime)
	relaySourceId := c.GetInstance().TowerId

	switch msg.EventTag {
	case client_model.ServerGoingDownEvent:
		c.Disconnect()
		return nil
	case client_model.BackupCompleteEvent:
		// Log the backup time, but don't return so the
		// message can be relayed to the web client
		err := pack.InstanceService.SetLastBackup(relaySourceId, sentTime)
		if err != nil {
			return err
		}

		if pack.InstanceService.GetLocal().IsCore() {
			// Also update the local core server's last backup time
			err = pack.InstanceService.SetLastBackup(pack.InstanceService.GetLocal().ServerId(), sentTime)
			if err != nil {
				return err
			}
		}
	case client_model.RemoteConnectionChangedEvent:
		return nil
	}

	msg.RelaySource = relaySourceId
	ctx.ClientService.Relay(msg)

	ctx.Logger.Debug().Msgf("Relaying message [%s] from server [%s]", msg.EventTag, relaySourceId)

	return nil
}

func onWebConnect(ctx context.RequestContext, c *client_model.WsClient) {
	if !pack.Loaded.Load() {
		c.PushWeblensEvent(client_model.StartupProgressEvent, client_model.WsC{"waitingOn": pack.GetStartupTasks()})
		return
	} else {
		c.PushWeblensEvent(client_model.WeblensLoadedEvent, client_model.WsC{"role": pack.InstanceService.GetLocal().GetRole()})
	}

	if pack.InstanceService.GetLocal().GetRole() == models.BackupServerRole {
		for _, backupTask := range pack.TaskService.GetTasksByJobName(models.BackupTask) {
			r := backupTask.GetResults()
			if len(r) == 0 {
				continue
			}

			c.PushTaskUpdate(backupTask, client_model.BackupProgressEvent, r)
		}
	}
}

func wsRecover(c client_model.Client) {
	name := ""
	if u := c.GetUser(); u != nil {
		name = u.GetUsername()
	}
	internal.RecoverPanic(fmt.Sprintf("[%s] websocket panic", name))
}
