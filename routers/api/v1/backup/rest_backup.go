package backup

import (
	"net/http"
	"time"

	"github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/jobs"
	"github.com/ethanrous/weblens/services/proxy"
	"github.com/pkg/errors"
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
//	@Param		serverId	path	string	true	"Server ID"
//
//	@Success	200
//	@Router		/tower/{serverId}/backup [post]
func LaunchBackup(ctx context.RequestContext) {
	serverId := ctx.Path("serverId")

	if serverId == "" {
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
		remote, err := tower_model.GetTowerById(ctx, serverId)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)

			return
		}

		_, err = proxy.NewCoreRequest(&remote, http.MethodPost, "/backup").Call()
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		client := ctx.ClientService.GetClientByTowerId(serverId)
		msg := websocket.WsResponseInfo{
			EventTag: "do_backup",
			Content:  websocket.WsData{"coreId": local.TowerId},
		}

		err = client.Send(msg)

		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}
	} else {
		core, err := tower_model.GetTowerById(ctx, serverId)
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
