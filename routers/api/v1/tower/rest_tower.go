package http

import (
	"errors"
	"net/http"

	"github.com/ethanrous/weblens/models/db"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/net"
	"github.com/ethanrous/weblens/modules/structs"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/reshape"
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
func GetServerInfo(ctx *context_service.RequestContext) {

	tower, err := tower_model.GetLocal(ctx)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)
		return
	}

	towerInfo := reshape.TowerToTowerInfo(tower)
	ctx.JSON(http.StatusOK, towerInfo)
}

// GetRemotes godoc
//
//	@ID			GetRemotes
//
//	@Summary	Get all remotes
//	@Tags Towers
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Success	200	{array}	structs.TowerInfo	"Tower Info"
//	@Router		/tower [get]
func GetRemotes(ctx *context_service.RequestContext) {
	remotes, err := tower_model.GetRemotes(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	// localRole := pack.InstanceService.GetLocal().GetRole()

	var serverInfos []structs.TowerInfo = []structs.TowerInfo{}
	for _, r := range remotes {
		// addr, _ := r.Address
		// client := pack.ClientService.GetClientByServerId(r.ServerId())
		// online := client != nil && client.Active.Load()
		//
		// var backupSize int64 = -1
		// if localRole == tower_model.BackupTowerRole {
		// 	backupSize = ctx.FileService.Size(r.TowerId)
		// }

		serverInfos = append(serverInfos, reshape.TowerToTowerInfo(r))
	}

	ctx.JSON(http.StatusOK, serverInfos)
}

// // AttachRemote godoc
// //
// //	@ID			CreateRemote
// //
// //	@Summary	Create a new remote
// //	@Tags Towers
// //
// //	@Security	SessionAuth[admin]
// //	@Security	ApiKeyAuth[admin]
// //
// //	@Param		request	body		structs.NewServerParams	true	"New Server Params"
// //	@Success	201		{object}	structs.TowerInfo			"New Server Info"
// //	@Success	400
// //	@Router		/tower [post]
// func attachRemote(w http.ResponseWriter, r *http.Request) {
// 	pack := getServices(r)
// 	local := pack.InstanceService.GetLocal()
//
// 	params, err := readCtxBody[structs.NewServerParams](w, r)
// 	if err != nil {
// 		return
// 	}
//
// 	pack.Log.Debug().Msgf("Attaching remote %s server %s with key %s", params.Role, params.Id, params.UsingKey)
//
// 	if params.Role == models.CoreServerRole {
// 		newCore, err := pack.InstanceService.AttachRemoteCore(params.CoreAddress, params.UsingKey)
// 		if SafeErrorAndExit(err, w) {
// 			return
// 		}
//
// 		mockJournal := mock.NewHollowJournalService()
// 		newTree, err := fileTree.NewFileTree(filepath.Join(pack.Cnf.DataRoot, newCore.ServerId()), newCore.ServerId(), mockJournal, false, pack.Log)
// 		if SafeErrorAndExit(err, w) {
// 			return
// 		}
//
// 		pack.FileService.AddTree(newTree)
//
// 		err = WebsocketToCore(newCore, pack)
// 		if SafeErrorAndExit(err, w) {
// 			return
// 		}
//
// 		coreInfo := structs.InstanceToServerInfo(newCore)
//
// 		writeJson(w, http.StatusCreated, coreInfo)
// 	} else if params.Role == models.BackupServerRole {
// 		newRemote := models.NewInstance(params.Id, params.Name, params.UsingKey, models.BackupServerRole, false, "", local.ServerId())
//
// 		err = pack.InstanceService.Add(newRemote)
// 		if err != nil {
// 			if errors.Is(err, werror.ErrKeyInUse) {
// 				w.WriteHeader(http.StatusConflict)
// 				return
// 			}
//
// 			pack.Log.Error().Stack().Err(err).Msg("Failed to add remote instance")
// 			writeJson(w, http.StatusInternalServerError, structs.WeblensErrorInfo{Error: err.Error()})
// 			return
// 		}
//
// 		err = pack.AccessService.SetKeyUsedBy(params.UsingKey, newRemote)
// 		if SafeErrorAndExit(err, w) {
// 			return
// 		}
//
// 		localInfo := structs.InstanceToServerInfo(pack.InstanceService.GetLocal())
//
// 		writeJson(w, http.StatusCreated, localInfo)
// 	} else {
// 		writeError(w, http.StatusBadRequest, werror.Errorf("'%s' is an invalid role. Must be 'core' or 'backup'", params.Role))
// 		return
// 	}
//
// 	jobs.RegisterJobs(pack.TaskService, pack.InstanceService.GetLocal().Role)
// }
//
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
//
// // DeleteRemote godoc
// //
// //	@ID			DeleteRemote
// //
// //	@Summary	Delete a remote
// //	@Tags Towers
// //
// //	@Security	SessionAuth[admin]
// //	@Security	ApiKeyAuth[admin]
// //
// //	@Param		serverId	path	string	true	"Server Id to delete"
// //	@Success	200
// //	@Success	400
// //	@Success	404
// //	@Router		/tower/{serverId} [delete]
// func removeRemote(w http.ResponseWriter, r *http.Request) {
// 	pack := getServices(r)
// 	remoteId := chi.URLParam(r, "serverId")
// 	if remoteId == "" {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
//
// 	remote := pack.InstanceService.GetByInstanceId(remoteId)
// 	if remote == nil {
// 		SafeErrorAndExit(werror.ErrNoInstance, w)
// 		return
// 	}
//
// 	err := pack.InstanceService.Del(remote.DbId)
// 	if SafeErrorAndExit(err, w) {
// 		return
// 	}
//
// 	if key := remote.GetUsingKey(); key != "" {
// 		err = pack.AccessService.SetKeyUsedBy(key, nil)
// 		if SafeErrorAndExit(err, w) {
// 			return
// 		}
// 	}
//
// 	w.WriteHeader(http.StatusOK)
// }

// InitializeTower godoc
//
//	@ID InitializeTower
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
func InitializeTower(ctx *context_service.RequestContext) {
	// Retrieve the local tower instance
	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	// Check if the server is already initialized
	if local.Role != tower_model.InitTowerRole {
		ctx.Error(http.StatusConflict, tower_model.ErrTowerAlreadyInitialized)
		return
	}

	// Read the initialization parameters from the request body
	initBody, err := net.ReadRequestBody[structs.InitServerParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)
		return
	}

	err = db.WithTransaction(ctx, func(sessionCtx context_mod.ContextZ) error {
		// Initialize the server based on the specified role
		switch tower_model.TowerRole(initBody.Role) {
		case tower_model.CoreTowerRole:
			if err := initializeCoreServer(sessionCtx, initBody); err != nil {
				return err
			}
		case tower_model.BackupTowerRole:
			if err := initializeBackupServer(sessionCtx, initBody); err != nil {
				return err
			}
		case tower_model.RestoreTowerRole:
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
	localInfo := reshape.TowerToTowerInfo(local)
	ctx.JSON(http.StatusCreated, localInfo)
}

func initializeCoreServer(ctx context_mod.ContextZ, initBody structs.InitServerParams) error {
	if initBody.Name == "" || initBody.Username == "" || initBody.Password == "" {
		return errors.New("missing required fields for core server initialization")
	}

	rqCtx := ctx.(*context_service.RequestContext)

	// Remove existing users and create a new owner
	user_model.DeleteAllUsers(rqCtx)

	owner := &user_model.User{
		Username:    initBody.Username,
		Password:    initBody.Password,
		DisplayName: initBody.FullName,
		UserPerms:   user_model.UserPermissionOwner,
		Activated:   true,
	}

	// Create user home directory
	err := rqCtx.FileService.CreateUserHome(ctx, owner)
	if err != nil {
		return err
	}

	err = user_model.CreateUser(ctx, owner)
	if err != nil {
		return err
	}

	local, err := tower_model.GetLocal(ctx)
	local.Role = tower_model.CoreTowerRole
	local.Name = initBody.Name
	local.Address = initBody.CoreAddress

	err = tower_model.UpdateTower(ctx, local)
	if err != nil {
		return err
	}

	rqCtx.Requester = owner
	err = rqCtx.SetSessionToken()
	if err != nil {
		rqCtx.Error(http.StatusInternalServerError, err)
	}

	return nil
}

func initializeBackupServer(ctx context_mod.ContextZ, initBody structs.InitServerParams) error {
	return errors.New("backup server initialization not implemented")

	// if initBody.Name == "" || initBody.Username == "" || initBody.Password == "" {
	// 	return errors.New("missing required fields for core server initialization")
	// }
	//
	// local, err := tower_model.GetLocal(ctx)
	// local.Role = tower_model.CoreTowerRole
	// local.Name = initBody.Name
	// local.Address = initBody.CoreAddress
	//
	// err = tower_model.UpdateTower(ctx, local)
	// if err != nil {
	// 	return err
	// }
	//
	// // Remove existing users and create a new owner
	// user_model.DeleteAllUsers(ctx)
	//
	// owner := &user_model.User{
	// 	Username:    initBody.Username,
	// 	Password:    initBody.Password,
	// 	DisplayName: initBody.FullName,
	// 	UserPerms:   user_model.UserPermissionOwner,
	// }
	//
	// err = user_model.CreateUser(ctx, owner)
	// if err != nil {
	// 	return err
	// }
	//
	// // Create user home directory
	// err = ctx.FileService.CreateUserHome(ctx, owner)
	// if err != nil {
	// 	return err
	// }
	//
	// ctx.Requester = owner
	// err = ctx.SetSessionToken()
	// if err != nil {
	// 	ctx.Error(http.StatusInternalServerError, err)
	// }
	//
	// return nil
}

// func initializeRestoreServer(ctx *context.RequestContext, initBody structs.InitServerParams) error {
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

//
// // ResetServer godoc
// //
// //	@ID			ResetServer
// //
// //	@Security	SessionAuth[admin]
// //	@Security	ApiKeyAuth[admin]
// //
// //	@Summary	Reset server
// //	@Tags Towers
// //	@Produce	json
// //
// //	@Success	202
// //	@Failure	404
// //	@Failure	500
// //	@Router		/tower/reset [post]
// func resetServer(w http.ResponseWriter, r *http.Request) {
// 	pack := getServices(r)
// 	u, err := getUserFromCtx(r, false)
// 	if SafeErrorAndExit(err, w) {
// 		return
// 	}
//
// 	if !u.IsOwner() {
// 		writeError(w, http.StatusForbidden, werror.ErrNotOwner)
// 		return
// 	}
//
// 	// Can't reset server if not initialized
// 	role := pack.InstanceService.GetLocal().GetRole()
// 	if role == models.InitServerRole {
// 		writeError(w, http.StatusNotFound, werror.ErrServerNotInitialized)
// 		return
// 	}
//
// 	err = pack.InstanceService.ResetAll()
// 	if SafeErrorAndExit(err, w) {
// 		return
// 	}
//
// 	err = pack.UserService.Del(u.GetUsername())
// 	if SafeErrorAndExit(err, w) {
// 		return
// 	}
//
// 	w.WriteHeader(http.StatusOK)
// }
//
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
