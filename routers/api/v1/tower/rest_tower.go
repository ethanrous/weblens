package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/ethanrous/weblens/models/db"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/net"
	"github.com/ethanrous/weblens/modules/startup"
	"github.com/ethanrous/weblens/modules/structs"
	"github.com/ethanrous/weblens/routers/api/v1/websocket"
	access_service "github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/reshape"
	tower_service "github.com/ethanrous/weblens/services/tower"
)

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

	serverInfos := make([]structs.TowerInfo, 0, len(remotes))
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
	params, err := net.ReadRequestBody[structs.NewServerParams](ctx.Req)
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

			core = reshape.ApiTowerInfoToTower(*towerInfo)
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
				TowerId:     params.Id,
				Name:        params.Name,
				IncomingKey: params.UsingKey,
				Role:        tower_model.RoleBackup,
				CreatedBy:   ctx.LocalTowerId,
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

	// reshape.TowerInfoToTower(params)
	//
	// pack.Log.Debug().Msgf("Attaching remote %s server %s with key %s", params.Role, params.Id, params.UsingKey)
	//
	// if params.Role == models.CoreServerRole {
	// 	newCore, err := pack.InstanceService.AttachRemoteCore(params.CoreAddress, params.UsingKey)
	// 	if SafeErrorAndExit(err, w) {
	// 		return
	// 	}
	//
	// 	mockJournal := mock.NewHollowJournalService()
	// 	newTree, err := fileTree.NewFileTree(filepath.Join(pack.Cnf.DataRoot, newCore.ServerId()), newCore.ServerId(), mockJournal, false, pack.Log)
	// 	if SafeErrorAndExit(err, w) {
	// 		return
	// 	}
	//
	// 	pack.FileService.AddTree(newTree)
	//
	// 	err = WebsocketToCore(newCore, pack)
	// 	if SafeErrorAndExit(err, w) {
	// 		return
	// 	}
	//
	// 	coreInfo := structs.InstanceToServerInfo(newCore)
	//
	// 	writeJson(w, http.StatusCreated, coreInfo)
	// } else if params.Role == models.BackupServerRole {
	// 	newRemote := models.NewInstance(params.Id, params.Name, params.UsingKey, models.BackupServerRole, false, "", local.ServerId())
	//
	// 	err = pack.InstanceService.Add(newRemote)
	// 	if err != nil {
	// 		if errors.Is(err, werror.ErrKeyInUse) {
	// 			w.WriteHeader(http.StatusConflict)
	// 			return
	// 		}
	//
	// 		pack.Log.Error().Stack().Err(err).Msg("Failed to add remote instance")
	// 		writeJson(w, http.StatusInternalServerError, structs.WeblensErrorInfo{Error: err.Error()})
	// 		return
	// 	}
	//
	// 	err = pack.AccessService.SetKeyUsedBy(params.UsingKey, newRemote)
	// 	if SafeErrorAndExit(err, w) {
	// 		return
	// 	}
	//
	// 	localInfo := structs.InstanceToServerInfo(pack.InstanceService.GetLocal())
	//
	// 	writeJson(w, http.StatusCreated, localInfo)
	// } else {
	// 	writeError(w, http.StatusBadRequest, werror.Errorf("'%s' is an invalid role. Must be 'core' or 'backup'", params.Role))
	// 	return
	// }
	//
	// jobs.RegisterJobs(pack.TaskService, pack.InstanceService.GetLocal().Role)
}

// // UpdateRemote godoc
// //
// //	@ID			UpdateRemote
// //
// //	@Summary	Update a remote
// //	@Tags Towers
// //
// //	@Security	SessionAuth[admin]
// //	@Security	ApiKeyAuth[admin]
// //
// //	@Param		serverId	path	string			true	"Server Id to update"
// //	@Param		request		body	structs.UpdateServerParams	true	"Server Params"
// //	@Success	200
// //	@Success	400
// //	@Success	404
// //	@Router		/tower/{serverId} [patch]
// func updateRemote(w http.ResponseWriter, r *http.Request) {
// 	pack := getServices(r)
// 	remoteId := chi.URLParam(r, "serverId")
// 	if remoteId == "" {
// 		writeError(w, http.StatusBadRequest, werror.ErrNoServerId)
// 		return
// 	}
//
// 	remote := pack.InstanceService.GetByInstanceId(remoteId)
// 	if remote == nil {
// 		SafeErrorAndExit(werror.ErrNoInstance, w)
// 		return
// 	}
//
// 	params, err := readCtxBody[structs.UpdateServerParams](w, r)
// 	if SafeErrorAndExit(err, w) {
// 		return
// 	}
//
// 	if params.Name != "" {
// 		remote.Name = params.Name
// 	}
//
// 	if params.UsingKey != "" {
// 		remote.UsingKey = params.UsingKey
// 	}
//
// 	if params.CoreAddress != "" {
// 		remote.Address = params.CoreAddress
// 	}
//
// 	err = pack.InstanceService.Update(remote)
// 	if SafeErrorAndExit(err, w) {
// 		return
// 	}
//
// 	w.WriteHeader(http.StatusOK)
// }

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
//	@Param		serverId	path	string	true	"Server Id to delete"
//	@Success	200
//	@Success	400
//	@Success	404
//	@Router		/tower/{serverId} [delete]
func DeleteRemote(ctx context_service.RequestContext) {
	remoteId := ctx.Path("serverId")

	_, err := tower_model.GetTowerById(ctx, remoteId)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	err = tower_model.DeleteTowerById(ctx, remoteId)
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
	if local.Role != tower_model.RoleInit {
		ctx.Error(http.StatusConflict, tower_model.ErrTowerAlreadyInitialized)

		return
	}

	// Read the initialization parameters from the request body
	initBody, err := net.ReadRequestBody[structs.InitServerParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	err = db.WithTransaction(ctx, func(sessionCtx context.Context) error {
		// Initialize the server based on the specified role
		switch tower_model.Role(initBody.Role) {
		case tower_model.RoleCore:
			if err := initializeCoreServer(sessionCtx, initBody); err != nil {
				return err
			}
		case tower_model.RoleBackup:
			if err := initializeBackupServer(sessionCtx, initBody); err != nil {
				return err
			}
		case tower_model.RoleRestore:
			{
				return errors.New("restore server initialization not implemented")
			}
		default:
			err = errors.New("invalid server role")

			return err
		}

		return nil
	})

	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	// Respond with the local server information
	localInfo := reshape.TowerToTowerInfo(ctx, local)
	ctx.JSON(http.StatusCreated, localInfo)
}

func newOwner(ctx context.Context, initBody structs.InitServerParams) (*user_model.User, error) {
	owner := &user_model.User{
		Username:    initBody.Username,
		Password:    initBody.Password,
		DisplayName: initBody.FullName,
		UserPerms:   user_model.UserPermissionOwner,
		Activated:   true,
	}

	rqCtx, ok := context_service.ReqFromContext(ctx)
	if !ok {
		return nil, errors.New("failed to get request context")
	}

	if tower_model.Role(initBody.Role) == tower_model.RoleCore {
		// Create user home directory
		err := rqCtx.FileService.CreateUserHome(ctx, owner)
		if err != nil {
			return nil, err
		}

		if owner.HomeId == "" {
			return nil, errors.New("failed to create user home directory")
		}
	}

	err := user_model.SaveUser(ctx, owner)
	if err != nil {
		return nil, err
	}

	rqCtx.Requester = owner

	err = access_service.SetSessionToken(rqCtx)
	if err != nil {
		rqCtx.Error(http.StatusInternalServerError, err)
	}

	return owner, nil
}

func initializeCoreServer(ctx context.Context, initBody structs.InitServerParams) error {
	if initBody.Name == "" || initBody.Username == "" || initBody.Password == "" {
		return errors.New("missing required fields for core server initialization")
	}

	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		return err
	}

	local.Role = tower_model.RoleCore
	local.Name = initBody.Name
	local.Address = initBody.CoreAddress

	err = tower_model.UpdateTower(ctx, &local)
	if err != nil {
		return err
	}

	cnf := config.GetConfig()
	cnf.InitRole = string(tower_model.RoleCore)
	err = startup.RunStartups(ctx, cnf)
	if err != nil {
		return err
	}

	_, err = newOwner(ctx, initBody)
	if err != nil {
		return err
	}

	return nil
}

func initializeBackupServer(ctx context.Context, initBody structs.InitServerParams) error {
	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		return err
	}

	local.Role = tower_model.RoleBackup
	local.Name = initBody.Name

	err = tower_model.UpdateTower(ctx, &local)
	if err != nil {
		return err
	}

	core := tower_model.Instance{
		Role:        tower_model.RoleCore,
		Address:     initBody.CoreAddress,
		OutgoingKey: initBody.CoreKey,
	}

	coreInfo, err := tower_service.Ping(ctx, core)
	if err != nil {
		return err
	}

	if tower_model.Role(coreInfo.GetRole()) != tower_model.RoleCore {
		return tower_model.ErrNotCore
	}

	core.TowerId = coreInfo.GetId()
	core.Name = coreInfo.GetName()

	err = tower_model.SaveTower(ctx, &core)
	if err != nil {
		return err
	}

	_, err = newOwner(ctx, initBody)
	if err != nil {
		return err
	}

	return nil
}

// func initializeRestoreServer(ctx context.RequestContext, initBody structs.InitServerParams) error {
// 	local := ctx.InstanceService.GetLocal()
// 	if local.Role == models.RestoreServerRole {
// 		return nil
// 	}
//
// 	err := ctx.AccessService.AddApiKey(initBody.UsingKeyInfo)
// 	if err != nil && !errors.Is(err, werror.ErrKeyAlreadyExists) {
// 		return err
// 	}
//
// 	local.SetRole(models.RestoreServerRole)
// 	ctx.Caster.PushWeblensEvent(models.RestoreStartedEvent)
//
// 	hasherFactory := func() fileTree.Hasher {
// 		return models.NewHasher(ctx.TaskService, ctx.Caster)
// 	}
//
// 	journalConfig := fileTree.JournalConfig{
// 		Collection:    ctx.Db.Collection("fileHistory"),
// 		ServerId:      initBody.LocalId,
// 		IgnoreLocal:   false,
// 		HasherFactory: hasherFactory,
// 		Logger:        ctx.Logger,
// 	}
//
// 	journal, err := fileTree.NewJournal(journalConfig)
// 	if err != nil {
// 		return err
// 	}
//
// 	usersTree, err := fileTree.NewFileTree(filepath.Join(ctx.Cnf.DataRoot, "users"), "USERS", journal, false, ctx.Logger)
// 	if err != nil {
// 		return err
// 	}
//
// 	ctx.FileService.AddTree(usersTree)
//
// 	return nil
// }

// ResetTower godoc
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

// // RestoreCore godoc
// //
// //	@ID			RestoreCore
// //
// //	@Security	SessionAuth[admin]
// //	@Security	ApiKeyAuth[admin]
// //
// //	@Summary	Restore target core server
// //	@Tags Towers
// //	@Produce	json
// //
// //	@Param		serverId	path	string					true	"Server Id"
// //	@Param		request		body	structs.RestoreCoreParams	true	"Restore Params"
// //
// //	@Success	202
// //	@Failure	404
// //	@Failure	500
// //	@Router		/tower/{serverId}/restore [post]
// func restoreToCore(w http.ResponseWriter, r *http.Request) {
// 	restoreInfo, err := readCtxBody[structs.RestoreCoreParams](w, r)
// 	if err != nil {
// 		return
// 	}
//
// 	pack := getServices(r)
//
// 	core := pack.InstanceService.GetByInstanceId(restoreInfo.ServerId)
// 	if core == nil {
// 		w.WriteHeader(http.StatusNotFound)
// 		return
// 	}
//
// 	err = core.SetAddress(restoreInfo.HostUrl)
// 	if SafeErrorAndExit(err, w) {
// 		return
// 	}
//
// 	meta := models.RestoreCoreMeta{
// 		Local: pack.InstanceService.GetLocal(),
// 		Core:  core,
// 		Pack:  pack,
// 	}
//
// 	_, err = pack.TaskService.DispatchJob(models.RestoreCoreTask, meta, nil)
// 	if SafeErrorAndExit(err, w) {
// 		return
// 	}
//
// 	w.WriteHeader(http.StatusAccepted)
// }
