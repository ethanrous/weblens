// Package backup provides REST API handlers for backup operations.
package backup

import (
	"net/http"
	"time"

	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/services/ctxservice"
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
//	@Param		serverID	path	string	true	"Server ID of the tower to back up"
//
//	@Success	200
//	@Router		/tower/{serverID}/backup [post]
func LaunchBackup(ctx ctxservice.RequestContext) {
	remoteTowerID := ctx.Path("serverID")

	if remoteTowerID == "" {
		ctx.Error(http.StatusBadRequest, wlerrors.New("Server ID is required"))

		return
	}

	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	if local.TowerID == remoteTowerID {
		ctx.Error(http.StatusBadRequest, wlerrors.New("Cannot back up to the same tower"))

		return
	}

	// If the local is core, we essentially forward the backup request to the specified backup server
	if local.IsCore() {
		remote, err := tower_model.GetTowerByID(ctx, remoteTowerID)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)

			return
		}

		// Create a proxy API client to communicate with the remote tower
		remoteClient, err := proxy.APIClientFromTower(remote)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		// Launch the backup on the remote tower
		_, err = remoteClient.TowersAPI.LaunchBackup(ctx, local.TowerID).Execute()
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		// Notify the client via WebSocket that the backup has been initiated
		client := ctx.ClientService.GetClientByTowerID(remoteTowerID)
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
		// If the local is a backup server, we launch the backup task here
		core, err := tower_model.GetTowerByID(ctx, remoteTowerID)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)

			return
		}

		// Create the backup job
		t, err := jobs.BackupOne(ctx, core)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		// Subscribe the client who made the request to the backup task for real-time updates
		err = ctx.ClientService.SubscribeToTask(ctx, ctx.Client(), t, time.Now())
		if err != nil {
			// Log the error but do not fail the request, as the backup task has been created successfully
			ctx.Log().Warn().Err(err).Msg("Failed to subscribe client to backup task")
		}
	}

	ctx.Status(http.StatusAccepted)
}
