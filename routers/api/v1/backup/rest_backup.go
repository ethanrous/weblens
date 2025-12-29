// Package backup provides REST API handlers for backup operations.
package backup

import (
	"net/http"
	"time"

	"github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/jobs"
	"github.com/ethanrous/weblens/services/proxy"
)

// LaunchBackup godoc
//
//	@ID			LaunchBackup
//
//	@Summary	Launch backup on a tower
//	@Tags		Towers
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Param		serverID	path	string	true	"Server ID"
//
//	@Success	200
//	@Router		/tower/{serverID}/backup [post]
func LaunchBackup(ctx context.RequestContext) {
	serverID := ctx.Path("serverID")

	if serverID == "" {
		ctx.Error(http.StatusBadRequest, errors.New("Server ID is required"))

		return
	}

	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	// If the local is core, we send the backup request to the specied backup server
	if local.IsCore() {
		remote, err := tower_model.GetTowerByID(ctx, serverID)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)

			return
		}

		_, err = proxy.NewCoreRequest(&remote, http.MethodPost, "/backup").Call()
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		client := ctx.ClientService.GetClientByTowerID(serverID)
		msg := websocket.WsResponseInfo{
			EventTag: "do_backup",
			Content:  websocket.WsData{"coreID": local.TowerID},
		}

		err = client.Send(msg)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}
	} else {
		core, err := tower_model.GetTowerByID(ctx, serverID)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)

			return
		}

		t, err := jobs.BackupOne(ctx, core)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		err = ctx.ClientService.SubscribeToTask(ctx, ctx.Client(), t.(*task.Task), time.Now())
		if err != nil {
			// TODO: Return situational error here because the task still was created, but the client was not subscribed
			ctx.Error(http.StatusInternalServerError, err)

			return
		}
	}

	ctx.Status(http.StatusOK)
}
