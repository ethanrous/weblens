package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	client_model "github.com/ethanrous/weblens/models/client"
	"github.com/ethanrous/weblens/models/job"
	share_model "github.com/ethanrous/weblens/models/share"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/websocket"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/notify"
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

func Connect(ctx *context_service.RequestContext) {
	// if ctx.Query(IsTowerQueryKey) == "true" {
	// } else {
	// }

	// if getServer {
	// 	server = getInstanceFromCtx(r)
	// 	if server == nil {
	// 		SafeErrorAndExit(errors.WithStack(errors.ErrNoServerInContext), w)
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

	client := ctx.ClientService.ClientConnect(ctx, conn, ctx.Requester)

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

	go wsMain(ctx, client)
}

func wsMain(ctx *context_service.RequestContext, c *client_model.WsClient) {
	defer ctx.ClientService.ClientDisconnect(ctx, c)
	defer wsRecover(ctx, c)
	var switchboard func(*context_service.RequestContext, []byte, *client_model.WsClient) error

	if c.GetUser() != nil {
		switchboard = wsWebClientSwitchboard
		onWebConnect(ctx, c)
	} else {
		switchboard = wsServerClientSwitchboard
		// if pack.Loaded.Load() {
		// 	pack.Log.Debug().Msgf("Server connected: %s -- local role is: %s", c.GetInstance().ServerId(), pack.InstanceService.GetLocal().GetRole())
		// 	c.PushWeblensEvent(models.WeblensLoadedEvent, models.WsC{"role": pack.InstanceService.GetLocal().GetRole()})
		// }
	}

	for {
		_, buf, err := c.ReadOne()
		if err != nil {
			break
		}
		go func() {
			defer wsRecover(ctx, c)
			err := switchboard(ctx, buf, c)
			if err != nil {
				c.Error(err)
			}
		}()
	}
}

func wsWebClientSwitchboard(ctx *context_service.RequestContext, msgBuf []byte, c *client_model.WsClient) error {
	// localTower, err := tower_model.GetLocal(ctx)
	// if err != nil {
	// 	return err
	// }

	// TODO: Verify that we want to ignore requests to the init server
	// if localTower.Role == tower_model.InitServerRole {
	// 	return errors.ErrServerNotInitialized
	// }

	var msg websocket_mod.WsRequestInfo
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		return errors.WithStack(err)
	}

	c.Log().Trace().Func(func(e *zerolog.Event) { e.Interface("websocket_msg_content", msg).Msg("Got wsmsg from client") })

	if msg.Action == websocket_mod.ReportError {
		c.Log().Error().Msgf("Web client reported error: %s", msg.Content)
		return nil
	}

	subInfo, err := newActionBody(msg)
	if err != nil {
		return err
	}

	switch subInfo.Action() {
	case websocket_mod.FolderSubscribe:
		{
			err = json.Unmarshal([]byte(msg.Content), &subInfo)
			if err != nil {
				return errors.WithStack(err)
			}

			var share *share_model.FileShare
			if subInfo.GetShareId() != "" {
				share, err = share_model.GetShareById(ctx, subInfo.GetShareId())
				if err != nil {
					return errors.WithStack(err)
				}
			}

			f, err := ctx.FileService.GetFileById(subInfo.GetKey())

			err = ctx.ClientService.SubscribeToFile(ctx, c, f, share, time.UnixMilli(msg.SentAt))
			if err != nil {
				return errors.WithStack(err)
			}
		}
	case websocket_mod.TaskSubscribe:
		key := subInfo.GetKey()
		if strings.HasPrefix(key, "TID#") {
			key = key[4:]

			t := ctx.TaskService.GetTask(key)
			err = ctx.ClientService.SubscribeToTask(ctx, c, t, time.UnixMilli(msg.SentAt))
			if err != nil {
				return errors.WithStack(err)
			}

			complete, _ := t.Status()
			if complete {
				notif := notify.NewTaskNotification(t, websocket_mod.TaskCompleteEvent, t.GetResults())
				err = c.Send(notif)
				if err != nil {
					return errors.WithStack(err)
				}

			}
		} else if strings.HasPrefix(key, "TT#") {
			// key = key[3:]
			//
			// _, _, err := ctx.ClientService.SubscribeToTask(c, key, client_model.TaskTypeSubscribe, time.Now(), nil)
			// if err != nil {
			// 	c.Error(err)
			// }
		}

	case websocket_mod.Unsubscribe:
		key := subInfo.GetKey()

		if strings.HasPrefix(key, "TID#") {
			key = key[4:]
		} else if strings.HasPrefix(key, "TT#") {
			key = key[3:]
		}

		err = ctx.ClientService.Unsubscribe(ctx, c, key, time.UnixMilli(msg.SentAt))
		if err != nil && !errors.Is(err, notify.ErrSubscriptionNotFound) {
			c.Error(err)
		} else if err != nil {
			c.Log().Warn().Msgf("Subscription [%s] not found in unsub task", key)
		}

	case websocket_mod.ScanDirectory:
		{
			local, err := tower_model.GetLocal(ctx)
			if err != nil {
				return err
			}

			if local.Role == tower_model.BackupTowerRole {
				return tower_model.ErrTowerIsBackup
			}

			var share *share_model.FileShare
			if subInfo.GetShareId() != "" {
				share, err = share_model.GetShareById(ctx, subInfo.GetShareId())
				if err != nil {
					return err
				}
			}

			folder, err := ctx.FileService.GetFileById(subInfo.GetKey())
			if err != nil {
				return errors.Errorf("could not find directory to scan")
			}

			if !auth.CanUserAccessFile(c.GetUser(), folder, share) {
				return errors.Errorf("user does not have access to directory")
			}

			meta := job.ScanMeta{
				File: folder,
			}

			var jobName string
			if folder.IsDir() {
				jobName = job.ScanDirectoryTask
			} else {
				jobName = job.ScanFileTask
			}

			t, err := ctx.TaskService.DispatchJob(ctx, jobName, meta, nil)
			if err != nil {
				return err
			}

			err = ctx.ClientService.SubscribeToTask(ctx, c, t, time.UnixMilli(msg.SentAt))

			if err != nil {
				return err
			}
		}

	case websocket_mod.CancelTask:
		{
			task := ctx.TaskService.GetTask(subInfo.GetKey())
			if task == nil {
				return errors.Errorf("could not find task T[%s] to cancel", subInfo.GetKey())
			}

			task.Cancel()
			notif := notify.NewTaskNotification(task, websocket_mod.TaskCanceledEvent, nil)
			ctx.Notify(notif)
		}

	default:
		{
			c.Error(fmt.Errorf("unknown websocket request type: %s", string(msg.Action)))
		}
	}

	return nil
}

func wsServerClientSwitchboard(ctx *context_service.RequestContext, msgBuf []byte, c *client_model.WsClient) error {
	defer wsRecover(ctx, c)

	var msg websocket.WsResponseInfo
	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		return err
	}

	if msg.SentTime == 0 {
		err := errors.Errorf("invalid sent time on relay message: [%s]", msg.EventTag)
		return err
	}

	sentTime := time.UnixMilli(msg.SentTime)
	relaySourceId := c.GetInstance().TowerId

	switch msg.EventTag {
	case websocket_mod.ServerGoingDownEvent:
		c.Disconnect()
		return nil
	case websocket_mod.BackupCompleteEvent:
		// Log the backup time, but don't return so the
		// message can be relayed to the web client
		err = tower_model.SetLastBackup(ctx, relaySourceId, sentTime)
		if err != nil {
			return err
		}

		local, err := tower_model.GetLocal(ctx)
		if err != nil {
			return err
		}
		if local.IsCore() {
			// Also update the local core server's last backup time
			err = tower_model.SetLastBackup(ctx, local.TowerId, sentTime)
			if err != nil {
				return err
			}
		}
	case websocket_mod.RemoteConnectionChangedEvent:
		return nil
	}

	msg.RelaySource = relaySourceId
	ctx.ClientService.Relay(msg)

	ctx.Logger.Debug().Msgf("Relaying message [%s] from server [%s]", msg.EventTag, relaySourceId)

	return nil
}

func onWebConnect(ctx *context_service.RequestContext, c *client_model.WsClient) error {
	// if !pack.Loaded.Load() {
	// 	c.PushWeblensEvent(client_model.StartupProgressEvent, client_model.WsC{"waitingOn": pack.GetStartupTasks()})
	// 	return
	// } else {
	// 	c.PushWeblensEvent(client_model.WeblensLoadedEvent, client_model.WsC{"role": pack.InstanceService.GetLocal().GetRole()})
	// }
	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		return err
	}

	if local.IsBackup() {
		for _, backupTask := range ctx.TaskService.GetTasksByJobName(job.BackupTask) {
			r := backupTask.GetResults()
			if len(r) == 0 {
				continue
			}
			notif := notify.NewTaskNotification(backupTask, websocket_mod.BackupProgressEvent, r)
			err = c.Send(notif)
		}
	}

	return nil
}

func wsRecover(ctx context.LoggerContext, c *client_model.WsClient) {
	e := recover()
	if e == nil {
		return
	}

	err, ok := e.(error)
	if !ok {
		err = errors.Errorf("%v", e)
	}
	name := ""
	if u := c.GetUser(); u != nil {
		name = u.GetUsername()
	}

	ctx.Log().Error().Stack().Err(errors.WithStack(err)).Msgf("Websocket connection for user %s panicked: %v", name, err)
}
