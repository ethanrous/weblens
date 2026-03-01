package tower

import (
	"errors"
	"net/http"

	"github.com/ethanrous/weblens/models/db"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/netwrk"
	"github.com/ethanrous/weblens/modules/wlog"
	"github.com/ethanrous/weblens/modules/wlstructs"
	"github.com/ethanrous/weblens/routers/api/v1/websocket"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/reshape"
	tower_service "github.com/ethanrous/weblens/services/tower"
	"github.com/rs/zerolog"
)

// GetServerHealthStatus godoc
//
//	@ID			GetServerHealthStatus
//
//	@Summary	Get server health status
//	@Tags		Towers
//	@Produce	json
//	@Success	200 {object}	structs.TowerHealth 	"Health status"
//	@Router		/health [get]
func GetServerHealthStatus(ctx context_service.RequestContext) {
	ctx.JSON(http.StatusOK, wlstructs.TowerHealth{
		Status: "Healthy",
	})
}

// GetServerInfo godoc
//
//	@ID			GetServerInfo
//
//	@Summary	Get server info
//	@Tags		Towers
//	@Produce	json
//	@Success	200	{object}	structs.TowerInfo	"Server info"
//	@Router		/info [get]
func GetServerInfo(ctx context_service.RequestContext) {
	tower, err := tower_model.GetLocal(ctx)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	towerInfo := reshape.TowerToTowerInfo(ctx, tower)

	if ctx.Doer().IsAdmin() {
		towerInfo.LogLevel = config.GetConfig().LogLevel.String()
	}

	ctx.JSON(http.StatusOK, towerInfo)
}

// GetRemotes godoc
//
//	@ID			GetRemotes
//
//	@Summary	Get all remotes
//	@Tags		Towers
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Success	200	{array}	structs.TowerInfo	"Tower Info"
//	@Router		/tower [get]
func GetRemotes(ctx context_service.RequestContext) {
	remotes, err := tower_model.GetRemotes(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	serverInfos := make([]wlstructs.TowerInfo, 0, len(remotes))
	for _, r := range remotes {
		serverInfos = append(serverInfos, reshape.TowerToTowerInfo(ctx, r))
	}

	ctx.JSON(http.StatusOK, serverInfos)
}

// AttachRemote godoc
//
//	@ID			CreateRemote
//
//	@Summary	Create a new remote
//	@Tags		Towers
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Param		request	body		structs.NewServerParams	true	"New Server Params"
//	@Success	201		{object}	structs.TowerInfo		"New Server Info"
//	@Success	400
//	@Router		/tower/remote [post]
func AttachRemote(ctx context_service.RequestContext) {
	params, err := netwrk.ReadRequestBody[wlstructs.NewServerParams](ctx.Req)
	if err != nil {
		return
	}

	newRole := tower_model.Role(params.Role)

	switch newRole {
	case tower_model.RoleCore:
		{
			core := tower_model.Instance{Address: params.CoreAddress, OutgoingKey: params.UsingKey}

			err = tower_service.AttachToCore(ctx, core)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err)

				return
			}

			towerInfo, err := tower_service.Ping(ctx, core)
			if err != nil {
				ctx.Error(http.StatusBadRequest, err)

				return
			}

			core = reshape.APITowerInfoToTower(*towerInfo)
			core.Address = params.CoreAddress
			core.OutgoingKey = params.UsingKey

			err = tower_model.SaveTower(ctx, &core)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err)

				return
			}

			err = websocket.ConnectCore(ctx, &core)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err)

				return
			}
		}
	case tower_model.RoleBackup:
		{
			newRemote := tower_model.Instance{
				TowerID:     params.ID,
				Name:        params.Name,
				IncomingKey: params.UsingKey,
				Role:        tower_model.RoleBackup,
				CreatedBy:   ctx.LocalTowerID,
			}

			err = tower_model.SaveTower(ctx, &newRemote)
			if err != nil {
				if db.IsAlreadyExists(err) {
					ctx.Error(http.StatusConflict, err)

					return
				}

				ctx.Error(http.StatusInternalServerError, err)

				return
			}
		}
	default:
		ctx.Error(http.StatusBadRequest, errors.New("invalid role"))

		return
	}

	ctx.Status(http.StatusCreated)
}

// DeleteRemote godoc
//
//	@ID			DeleteRemote
//
//	@Summary	Delete a remote
//	@Tags		Towers
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Param		serverID	path	string	true	"Server ID to delete"
//	@Success	200
//	@Success	400
//	@Success	404
//	@Router		/tower/{serverID} [delete]
func DeleteRemote(ctx context_service.RequestContext) {
	remoteID := ctx.Path("serverID")

	_, err := tower_model.GetTowerByID(ctx, remoteID)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	err = tower_model.DeleteTowerByID(ctx, remoteID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.Status(http.StatusOK)
}

// InitializeTower godoc
//
//	@ID	InitializeTower
//
//	@Security
//
//	@Summary	Initialize the target server
//	@Tags		Towers
//	@Produce	json
//
//	@Param		request	body	structs.InitServerParams	true	"Server initialization body"
//
//	@Success	200		{array}	structs.TowerInfo			"New server info"
//	@Failure	404
//	@Failure	500
//	@Router		/tower/init [post]
func InitializeTower(ctx context_service.RequestContext) {
	// Retrieve the local tower instance
	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	// Check if the server is already initialized
	if local.Role != tower_model.RoleUninitialized {
		ctx.Error(http.StatusConflict, tower_model.ErrTowerAlreadyInitialized)

		return
	}

	// Read the initialization parameters from the request body
	initBody, err := netwrk.ReadRequestBody[wlstructs.InitServerParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	// Initialize the server based on the specified role
	// Note: No transaction wrapper - initialization is idempotent and can be retried
	switch tower_model.Role(initBody.Role) {
	case tower_model.RoleCore:
		err = tower_service.InitializeCoreServer(ctx, initBody, config.GetConfig())
	case tower_model.RoleBackup:
		err = tower_service.InitializeBackupServer(ctx, initBody, config.GetConfig())
	case tower_model.RoleRestore:
		err = errors.New("restore server initialization not implemented")
	default:
		err = errors.New("invalid server role")
	}

	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	// Respond with the local server information
	localInfo := reshape.TowerToTowerInfo(ctx, local)
	ctx.JSON(http.StatusCreated, localInfo)
}

// ResetServer godoc
//
//	@ID			ResetTower
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Summary	Reset tower
//	@Tags		Towers
//	@Produce	json
//
//	@Success	202
//	@Failure	404
//	@Failure	500
//	@Router		/tower/reset [post]
func ResetServer(ctx context_service.RequestContext) {
	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	// Can't reset server if not initialized
	if local.Role == tower_model.RoleRestore {
		ctx.Error(http.StatusBadRequest, tower_model.ErrTowerNotInitialized)

		return
	}

	err = tower_service.ResetTower(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.Status(http.StatusOK)
}

// EnableTraceLogging godoc
//
//	@ID			EnableTraceLogging
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Summary	Enable trace logging
//	@Tags		Towers
//	@Produce	json
//
//	@Success	200
//	@Router		/tower/trace [post]
func EnableTraceLogging(ctx context_service.RequestContext) {
	wlog.SetLogLevel(zerolog.TraceLevel)
	ctx.Status(http.StatusOK)
}
