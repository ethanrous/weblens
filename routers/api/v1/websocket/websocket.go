package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	client_model "github.com/ethanrous/weblens/models/client"
	"github.com/ethanrous/weblens/models/job"
	share_model "github.com/ethanrous/weblens/models/share"
	task_model "github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/websocket"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/ethanrous/weblens/services/reshape"
	gorilla "github.com/gorilla/websocket"
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

func Connect(ctx context_service.RequestContext) {
	conn, err := upgrader.Upgrade(ctx.W, ctx.Req, nil)
	if err != nil {
		ctx.Log().Error().Err(err).Msg("Failed to upgrade connection to websocket")

		return
	}

	client, err := ctx.ClientService.ClientConnect(ctx, conn, ctx.Requester)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to connect client")

		return
	}

	go wsMain(ctx, client)
}

func wsMain(ctx context_service.RequestContext, c *client_model.WsClient) {
	defer wsRecover(ctx, c)
	defer ctx.ClientService.ClientDisconnect(ctx, c)

	var switchboard func(context_service.RequestContext, []byte, *client_model.WsClient) error

	if c.GetUser() != nil {
		switchboard = wsWebClientSwitchboard

		err := onWebConnect(ctx, c)
		if err != nil {
			ctx.Log().Err(err).Msg("Failed to handle web client connect")

			return
		}
	} else {
		switchboard = wsServerClientSwitchboard
	}

	defer context.AfterFunc(ctx, func() {
		c.Disconnect()
	})()

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

func wsWebClientSwitchboard(ctx context_service.RequestContext, msgBuf []byte, c *client_model.WsClient) error {
	// localTower, err := tower_model.GetLocal(ctx)
	// if err != nil {
	// 	return err
	// }

	// TODO: Verify that we want to ignore requests to the init server
	// if localTower.Role == tower_model.InitServerRole {
	// 	return errors.ErrServerNotInitialized
	// }

	var msg websocket_mod.WsResponseInfo

	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		ctx.Log().Error().Msgf("Failed to unmarshal websocket message: %s", string(msgBuf))

		return errors.WithStack(err)
	}

	c.Log().Trace().Func(func(e *zerolog.Event) {
		e.Interface("websocket_msg_content", msg).Msgf("Got [%s/%s] wsmsg from client", msg.Action, msg.BroadcastType)
	})

	switch msg.Action {
	case websocket_mod.ReportError:
		{
			c.Log().Error().Msgf("Web client reported error: %s", msg.Content)

			return nil
		}
	case websocket_mod.ActionSubscribe:
		{
			subscription := reshape.GetSubscribeInfo(msg)
			if subscription.SubscriptionId == "" {
				return errors.Errorf("subscribe request missing subscription id")
			}

			switch subscription.Type {
			case websocket_mod.FolderSubscribe:
				{
					var share *share_model.FileShare

					shareId := share_model.ShareIdFromString(subscription.ShareId)
					if !shareId.IsZero() {
						share, err = share_model.GetShareById(ctx, shareId)
						if err != nil {
							return errors.WithStack(err)
						}
					}

					f, err := ctx.FileService.GetFileById(ctx, subscription.SubscriptionId)
					if err != nil {
						return errors.WithStack(err)
					}

					err = ctx.ClientService.SubscribeToFile(ctx, c, f, share, time.UnixMilli(msg.SentTime))
					if err != nil {
						return errors.WithStack(err)
					}
				}
			case websocket_mod.TaskSubscribe:
				key := subscription.SubscriptionId

				t := ctx.TaskService.GetTask(key)
				if t == nil {
					return errors.Errorf("could not find task T[%s] to subscribe", key)
				}

				ctx.Log().Debug().Msgf("Subscribing to task [%s]", key)

				err = ctx.ClientService.SubscribeToTask(ctx, c, t, time.UnixMilli(msg.SentTime))
				if err != nil && !errors.Is(err, task_model.ErrTaskAlreadyComplete) {
					return errors.Errorf("could not subscribe to task T[%s]: %w", key, err)
				} else if err != nil {
					notif := notify.NewTaskNotification(t, websocket_mod.TaskCompleteEvent, t.GetResults())

					err = c.Send(notif)
					if err != nil {
						return errors.WithStack(err)
					}
				}
			case websocket_mod.TaskTypeSubscribe:
				ctx.Log().Error().Msgf("Task type subscription not implemented: %s", subscription.SubscriptionId)
			default:
				return errors.Errorf("unknown subscription type: %s", msg.BroadcastType)
			}
		}

	case websocket_mod.ActionUnsubscribe:
		subscription := reshape.GetSubscribeInfo(msg)

		if strings.HasPrefix(subscription.SubscriptionId, "TID#") {
			subscription.SubscriptionId = subscription.SubscriptionId[4:]
		} else if strings.HasPrefix(subscription.SubscriptionId, "TT#") {
			subscription.SubscriptionId = subscription.SubscriptionId[3:]
		}

		err = ctx.ClientService.Unsubscribe(ctx, c, subscription.SubscriptionId, time.UnixMilli(msg.SentTime))
		if err != nil && !errors.Is(err, notify.ErrSubscriptionNotFound) {
			c.Error(err)
		} else if err != nil {
			c.Log().Warn().Msgf("Subscription [%s] not found in websocket unsub", subscription.SubscriptionId)
		}

	case websocket_mod.ScanDirectory:
		{
			scanInfo := reshape.GetScanInfo(msg)

			local, err := tower_model.GetLocal(ctx)
			if err != nil {
				return err
			}

			if local.Role == tower_model.RoleBackup {
				return tower_model.ErrTowerIsBackup
			}

			var share *share_model.FileShare

			if scanInfo.ShareId != "" {
				shareId := share_model.ShareIdFromString(ctx.Path("shareId"))

				share, err = share_model.GetShareById(ctx, shareId)
				if err != nil {
					return err
				}
			}

			folder, err := ctx.FileService.GetFileById(ctx, scanInfo.FileId)
			if err != nil {
				ctx.Log().Error().Stack().Err(err).Msgf("Failed to get file by id: %s", scanInfo.FileId)

				return errors.Errorf("could not find directory to scan: %w", err)
			}

			if _, err = auth.CanUserAccessFile(ctx, c.GetUser(), folder, share); err != nil {
				return err
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

			err = ctx.ClientService.SubscribeToTask(ctx, c, t.(*task_model.Task), time.UnixMilli(msg.SentTime))
			if err != nil {
				return err
			}
		}

	case websocket_mod.CancelTask:
		{
			cancelInfo := reshape.GetCancelInfo(msg)
			task := ctx.TaskService.GetTask(cancelInfo.TaskId)
			if task == nil {
				return errors.Errorf("could not find task T[%s] to cancel", cancelInfo.TaskId)
			}

			task.Cancel()
			notif := notify.NewTaskNotification(task, websocket_mod.TaskCanceledEvent, nil)
			ctx.Notify(ctx, notif)
		}

	case "":
	default:
		{
			c.Error(fmt.Errorf("unknown websocket request type: %s", string(msg.Action)))
		}
	}

	return nil
}

func wsServerClientSwitchboard(ctx context_service.RequestContext, msgBuf []byte, c *client_model.WsClient) error {
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

	ctx.Log().Debug().Msgf("Relaying message [%s] from server [%s]", msg.EventTag, relaySourceId)

	return nil
}

func onWebConnect(ctx context_service.RequestContext, c *client_model.WsClient) error {
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

func wsRecover(ctx context_mod.LoggerContext, c *client_model.WsClient) {
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
