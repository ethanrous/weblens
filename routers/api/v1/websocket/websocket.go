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
	"github.com/ethanrous/weblens/modules/websocket"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/ethanrous/weblens/services/reshape"
	gorilla "github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// IsTowerQueryKey is the query parameter key to indicate a tower connection.
const IsTowerQueryKey = "isTower"

// WsAuthorize represents the authorization message for websocket connections.
type WsAuthorize struct {
	Auth string `json:"auth"`
}

var upgrader = gorilla.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

// Connect upgrades an HTTP connection to a websocket connection.
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
	defer ctx.ClientService.ClientDisconnect(ctx, c) //nolint:errcheck

	var switchboard func(context_service.RequestContext, []byte, *client_model.WsClient) error

	if c.GetUser() != nil {
		switchboard = wsWebClientSwitchboard

		err := onWebConnect(ctx, c)
		if err != nil {
			ctx.Log().Err(err).Msg("Failed to handle web client connect")

			return
		}
	} else {
		switchboard = wsTowerClientSwitchboard
	}

	defer context.AfterFunc(ctx, func() {
		err := ctx.ClientService.ClientDisconnect(ctx, c)
		if err != nil {
			ctx.Log().Error().Stack().Err(err).Msg("Failed to disconnect websocket client")
		}
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

func handleActionSubscribe(msg websocket_mod.WsResponseInfo, ctx context_service.RequestContext, c *client_model.WsClient) error {
	subscription := reshape.GetSubscribeInfo(msg)
	if subscription.SubscriptionID == "" {
		return wlerrors.Errorf("subscribe request missing subscription id")
	}

	switch subscription.Type {
	case websocket_mod.FolderSubscribe:
		{
			var share *share_model.FileShare

			var err error

			shareID := share_model.IDFromString(subscription.ShareID)
			if !shareID.IsZero() {
				share, err = share_model.GetShareByID(ctx, shareID)
				if err != nil {
					return wlerrors.WithStack(err)
				}
			}

			file, err := ctx.FileService.GetFileByID(ctx, subscription.SubscriptionID)
			if err != nil {
				return wlerrors.WithStack(err)
			}

			if _, err = auth.CanUserAccessFile(ctx, c.GetUser(), file, share, share_model.SharePermissionView); err != nil {
				return err
			}

			err = ctx.ClientService.SubscribeToFile(ctx, c, file, time.UnixMilli(msg.SentTime))
			if err != nil {
				return wlerrors.WithStack(err)
			}

			fInfo, err := reshape.WeblensFileToFileInfo(ctx, file)
			if err != nil {
				return err
			}

			fileNotif := notify.NewFileNotification(ctx, fInfo, websocket_mod.FileUpdatedEvent)
			ctx.ClientService.Notify(ctx, fileNotif...)
		}
	case websocket_mod.TaskSubscribe:
		key := subscription.SubscriptionID

		t := ctx.TaskService.GetTask(key)
		if t == nil {
			return wlerrors.Errorf("could not find task T[%s] to subscribe", key)
		}

		ctx.Log().Debug().Msgf("Subscribing to task [%s]", key)

		err := ctx.ClientService.SubscribeToTask(ctx, c, t, time.UnixMilli(msg.SentTime))
		if err != nil && !wlerrors.Is(err, task_model.ErrTaskAlreadyComplete) {
			return wlerrors.Errorf("could not subscribe to task T[%s]: %w", key, err)
		} else if err != nil {
			notif := notify.NewTaskNotification(t, websocket_mod.TaskCompleteEvent, t.GetResults())

			err = c.Send(notif)
			if err != nil {
				return wlerrors.WithStack(err)
			}
		}
	case websocket_mod.TaskTypeSubscribe:
		ctx.Log().Error().Msgf("Task type subscription not implemented: %s", subscription.SubscriptionID)
	default:
		return wlerrors.Errorf("unknown subscription type: %s", msg.BroadcastType)
	}

	return nil
}

func handleScanDirectory(ctx context_service.RequestContext, msg websocket_mod.WsResponseInfo, c *client_model.WsClient) error {
	scanInfo := reshape.GetScanInfo(msg)

	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		return err
	}

	if local.Role == tower_model.RoleBackup {
		return tower_model.ErrTowerIsBackup
	}

	var share *share_model.FileShare

	if scanInfo.ShareID != "" {
		shareID := share_model.IDFromString(ctx.Path("shareID"))

		share, err = share_model.GetShareByID(ctx, shareID)
		if err != nil {
			return err
		}
	}

	folder, err := ctx.FileService.GetFileByID(ctx, scanInfo.FileID)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msgf("Failed to get file by id: %s", scanInfo.FileID)

		return wlerrors.Errorf("could not find directory to scan: %w", err)
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

	err = ctx.ClientService.SubscribeToTask(ctx, c, t, time.UnixMilli(msg.SentTime))
	if err != nil {
		return err
	}

	return nil
}

func wsWebClientSwitchboard(ctx context_service.RequestContext, msgBuf []byte, c *client_model.WsClient) error {
	var msg websocket_mod.WsResponseInfo

	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		ctx.Log().Error().Msgf("Failed to unmarshal websocket message: %s", string(msgBuf))

		return wlerrors.WithStack(err)
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
			err = handleActionSubscribe(msg, ctx, c)
			if err != nil {
				return err
			}
		}

	case websocket_mod.ActionUnsubscribe:
		subscription := reshape.GetSubscribeInfo(msg)

		if strings.HasPrefix(subscription.SubscriptionID, "TID#") {
			subscription.SubscriptionID = subscription.SubscriptionID[4:]
		} else if strings.HasPrefix(subscription.SubscriptionID, "TT#") {
			subscription.SubscriptionID = subscription.SubscriptionID[3:]
		}

		err = ctx.ClientService.Unsubscribe(ctx, c, subscription.SubscriptionID, time.UnixMilli(msg.SentTime))
		if err != nil && !wlerrors.Is(err, notify.ErrSubscriptionNotFound) {
			c.Error(err)
		} else if err != nil {
			c.Log().Warn().Msgf("Subscription [%s] not found in websocket unsub", subscription.SubscriptionID)
		}

	case websocket_mod.ScanDirectory:
		{
			err = handleScanDirectory(ctx, msg, c)
			if err != nil {
				return err
			}
		}

	case websocket_mod.CancelTask:
		{
			cancelInfo := reshape.GetCancelInfo(msg)

			task := ctx.TaskService.GetTask(cancelInfo.TaskID)
			if task == nil {
				return wlerrors.Errorf("could not find task T[%s] to cancel", cancelInfo.TaskID)
			}

			notif := notify.NewTaskNotification(task, websocket_mod.TaskCanceledEvent, nil)
			ctx.Notify(ctx, notif)
			<-notif.Sent

			task.Cancel()
		}
	case websocket_mod.RefreshTower:
		{
			ctx.Log().Debug().Msgf("Received refresh tower request from web client %+v", msg)
			towerInfo := reshape.GetTowerInfo(msg)

			remoteTower, err := tower_model.GetTowerByID(ctx, towerInfo.TowerID)
			if err != nil {
				return wlerrors.Errorf("could not find tower T[%s] to refresh: %w", towerInfo.TowerID, err)
			}

			ok := ctx.ClientService.PushTowerUpdate(remoteTower)
			if !ok {
				ctx.Log().Warn().Msgf("Connection to tower [%s] was not retried", remoteTower.TowerID)
			}
		}

	case "":
	default:
		{
			c.Error(fmt.Errorf("unknown websocket request type: %s", string(msg.Action)))
		}
	}

	return nil
}

func wsTowerClientSwitchboard(ctx context_service.RequestContext, msgBuf []byte, c *client_model.WsClient) error {
	defer wsRecover(ctx, c)

	var msg websocket.WsResponseInfo

	err := json.Unmarshal(msgBuf, &msg)
	if err != nil {
		return err
	}

	if msg.SentTime == 0 {
		err := wlerrors.Errorf("invalid sent time on relay message: [%s]", msg.EventTag)

		return err
	}

	sentTime := time.UnixMilli(msg.SentTime)
	relaySourceID := c.GetInstance().TowerID

	switch msg.EventTag {
	case websocket_mod.ServerGoingDownEvent:
		err = c.Disconnect()
		if err != nil {
			return err
		}

		return nil
	case websocket_mod.BackupCompleteEvent:
		backupSize := int64(0)

		ctx.Log().Trace().Msgf("Received backup complete event from server [%s]: %+v", relaySourceID, msg.Content)

		// Log the backup time and size, but don't return so the
		// message can be relayed to the web client
		err = tower_model.SetLastBackup(ctx, relaySourceID, sentTime, backupSize)
		if err != nil {
			return err
		}

		local, err := tower_model.GetLocal(ctx)
		if err != nil {
			return err
		}

		if local.IsCore() {
			// Also update the local core server's last backup time
			err = tower_model.SetLastBackup(ctx, local.TowerID, sentTime, backupSize)
			if err != nil {
				return err
			}
		}
	case websocket_mod.RemoteConnectionChangedEvent:
		return nil
	}

	msg.RelaySource = relaySourceID
	ctx.ClientService.Relay(msg)

	ctx.Log().Debug().Msgf("Relaying message [%s] from server [%s]", msg.EventTag, relaySourceID)

	return nil
}

func onWebConnect(ctx context_service.RequestContext, c *client_model.WsClient) error {
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

			err := c.Send(notif)
			if err != nil {
				return err
			}
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
		err = wlerrors.Errorf("%v", e)
	}

	name := ""
	if u := c.GetUser(); u != nil {
		name = u.GetUsername()
	}

	ctx.Log().Error().Stack().Err(wlerrors.WithStack(err)).Msgf("Websocket connection for user %s panicked: %v", name, err)
}
