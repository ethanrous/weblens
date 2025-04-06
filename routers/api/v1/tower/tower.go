package http

import (
	"net/http"

	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/reshape"
)

// GetServerInfo godoc
//
//	@ID			GetServerInfo
//
//	@Summary	Get server info
//	@Tags		Servers
//	@Produce	json
//	@Success	200	{object}	rest.ServerInfo	"Server info"
//	@Router		/info [get]
func GetServerInfo(ctx *context.RequestContext) {
	tower, err := tower_model.GetTowerById(ctx, ctx.LocalTowerId)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)
		return
	}

	towerInfo := reshape.TowerToTowerInfo(tower)
	ctx.JSON(http.StatusOK, towerInfo)
}

// // GetRemotes godoc
// //
// //	@ID			GetRemotes
// //
// //	@Summary	Get all remotes
// //	@Tags		Servers
// //
// //	@Security	SessionAuth[admin]
// //	@Security	ApiKeyAuth[admin]
// //
// //	@Success	200	{array}	rest.ServerInfo	"Server Info"
// //	@Router		/servers [get]
// func getRemotes(w http.ResponseWriter, r *http.Request) {
// 	pack := getServices(r)
//
// 	remotes := pack.InstanceService.GetRemotes()
// 	localRole := pack.InstanceService.GetLocal().GetRole()
//
// 	var serverInfos []rest.ServerInfo = []rest.ServerInfo{}
// 	for _, srv := range remotes {
// 		addr, _ := srv.GetAddress()
// 		client := pack.ClientService.GetClientByServerId(srv.ServerId())
// 		online := client != nil && client.Active.Load()
//
// 		var backupSize int64 = -1
// 		if localRole == models.BackupServerRole {
// 			backupSize = pack.FileService.Size(srv.ServerId())
// 		}
// 		serverInfos = append(serverInfos, rest.ServerInfo{
// 			Id:           srv.ServerId(),
// 			Name:         srv.GetName(),
// 			UsingKey:     srv.GetUsingKey(),
// 			Role:         srv.GetRole(),
// 			IsThisServer: srv.IsLocal(),
// 			Address:      addr,
// 			Online:       online,
// 			ReportedRole: srv.GetReportedRole(),
// 			LastBackup:   srv.LastBackup,
// 			BackupSize:   backupSize,
// 		})
// 	}
//
// 	writeJson(w, http.StatusOK, serverInfos)
// }
//
// // AttachRemote godoc
// //
// //	@ID			CreateRemote
// //
// //	@Summary	Create a new remote
// //	@Tags		Servers
// //
// //	@Security	SessionAuth[admin]
// //	@Security	ApiKeyAuth[admin]
// //
// //	@Param		request	body		rest.NewServerParams	true	"New Server Params"
// //	@Success	201		{object}	rest.ServerInfo			"New Server Info"
// //	@Success	400
// //	@Router		/servers [post]
// func attachRemote(w http.ResponseWriter, r *http.Request) {
// 	pack := getServices(r)
// 	local := pack.InstanceService.GetLocal()
//
// 	params, err := readCtxBody[rest.NewServerParams](w, r)
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
// 		coreInfo := rest.InstanceToServerInfo(newCore)
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
// 			writeJson(w, http.StatusInternalServerError, rest.WeblensErrorInfo{Error: err.Error()})
// 			return
// 		}
//
// 		err = pack.AccessService.SetKeyUsedBy(params.UsingKey, newRemote)
// 		if SafeErrorAndExit(err, w) {
// 			return
// 		}
//
// 		localInfo := rest.InstanceToServerInfo(pack.InstanceService.GetLocal())
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
// //	@Tags		Servers
// //
// //	@Security	SessionAuth[admin]
// //	@Security	ApiKeyAuth[admin]
// //
// //	@Param		serverId	path	string			true	"Server Id to update"
// //	@Param		request		body	rest.UpdateServerParams	true	"Server Params"
// //	@Success	200
// //	@Success	400
// //	@Success	404
// //	@Router		/servers/{serverId} [patch]
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
// 	params, err := readCtxBody[rest.UpdateServerParams](w, r)
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
// //	@Tags		Servers
// //
// //	@Security	SessionAuth[admin]
// //	@Security	ApiKeyAuth[admin]
// //
// //	@Param		serverId	path	string	true	"Server Id to delete"
// //	@Success	200
// //	@Success	400
// //	@Success	404
// //	@Router		/servers/{serverId} [delete]
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
//
// // InitializeServer godoc
// //
// //	@ID	InitializeServer
// //
// //	@Security
// //
// //	@Summary	Initialize the target server
// //	@Tags		Servers
// //	@Produce	json
// //
// //	@Param		request	body	rest.InitServerParams	true	"Server initialization body"
// //
// //	@Success	200		{array}	rest.ServerInfo			"New server info"
// //	@Failure	404
// //	@Failure	500
// //	@Router		/servers/init [post]
// func initializeServer(w http.ResponseWriter, r *http.Request) {
// 	pack := getServices(r)
// 	// Can't init server if already initialized
// 	role := pack.InstanceService.GetLocal().GetRole()
// 	if role != models.InitServerRole {
// 		writeError(w, http.StatusConflict, werror.ErrServerAlreadyInitialized)
// 		return
// 	}
//
// 	initBody, err := readCtxBody[rest.InitServerParams](w, r)
// 	if SafeErrorAndExit(err, w) {
// 		return
// 	}
//
// 	if initBody.Role == models.CoreServerRole {
// 		if initBody.Name == "" || initBody.Username == "" || initBody.Password == "" {
// 			writeError(w, http.StatusBadRequest, werror.Errorf("missing required fields"))
// 			return
// 		}
//
// 		err = pack.InstanceService.InitCore(initBody.Name)
// 		if SafeErrorAndExit(err, w) {
// 			return
// 		}
//
// 		users, err := pack.UserService.GetAll()
// 		if SafeErrorAndExit(err, w) {
// 			return
// 		}
//
// 		for u := range users {
// 			err = pack.UserService.Del(u.GetUsername())
// 			if err != nil {
// 				pack.Log.Error().Stack().Err(err).Msg("")
// 			}
// 		}
//
// 		owner, err := pack.UserService.CreateOwner(initBody.Username, initBody.Password, initBody.FullName)
// 		if SafeErrorAndExit(err, w) {
// 			return
// 		}
//
// 		err = pack.FileService.CreateUserHome(owner)
// 		if SafeErrorAndExit(err, w) {
// 			return
// 		}
//
// 		token, expires, err := pack.AccessService.GenerateJwtToken(owner)
// 		if SafeErrorAndExit(err, w) {
// 			return
// 		}
//
// 		cookie := fmt.Sprintf("%s=%s; expires=%s;", SessionTokenCookie, token, expires.Format(time.RFC1123))
// 		w.Header().Set("Set-Cookie", cookie)
//
// 		// go pack.Server.Restart(false)
// 	} else if initBody.Role == models.BackupServerRole {
// 		if initBody.Name == "" {
// 			w.WriteHeader(http.StatusBadRequest)
// 			return
// 		}
// 		if initBody.CoreAddress[len(initBody.CoreAddress)-1:] != "/" {
// 			initBody.CoreAddress += "/"
// 		}
//
// 		// Initialize the server as backup
// 		err = service.InitBackup(pack, initBody.Name, initBody.CoreAddress, initBody.CoreKey)
// 		if err != nil {
// 			pack.InstanceService.GetLocal().SetRole(models.InitServerRole)
// 			pack.Log.Error().Stack().Err(err).Msg("")
// 			w.WriteHeader(http.StatusBadRequest)
// 			return
// 		}
//
// 		writeJson(w, http.StatusCreated, pack.InstanceService.GetLocal())
//
// 		// go pack.Server.Restart()
// 		return
// 	} else if initBody.Role == models.RestoreServerRole {
// 		local := pack.InstanceService.GetLocal()
// 		if local.Role == models.RestoreServerRole {
// 			w.WriteHeader(http.StatusOK)
// 			return
// 		}
//
// 		err = pack.AccessService.AddApiKey(initBody.UsingKeyInfo)
// 		if err != nil && !errors.Is(err, werror.ErrKeyAlreadyExists) {
// 			SafeErrorAndExit(err, w)
// 			return
// 		}
//
// 		local.SetRole(models.RestoreServerRole)
// 		pack.Caster.PushWeblensEvent(models.RestoreStartedEvent)
//
// 		hasherFactory := func() fileTree.Hasher {
// 			return models.NewHasher(pack.TaskService, pack.Caster)
// 		}
//
// 		journalConfig := fileTree.JournalConfig{
// 			Collection:    pack.Db.Collection("fileHistory"),
// 			ServerId:      initBody.LocalId,
// 			IgnoreLocal:   false,
// 			HasherFactory: hasherFactory,
// 			Logger:        pack.Log,
// 		}
//
// 		journal, err := fileTree.NewJournal(journalConfig)
// 		if SafeErrorAndExit(err, w) {
// 			return
// 		}
// 		usersTree, err := fileTree.NewFileTree(filepath.Join(pack.Cnf.DataRoot, "users"), "USERS", journal, false, pack.Log)
// 		if SafeErrorAndExit(err, w) {
// 			return
// 		}
// 		pack.FileService.AddTree(usersTree)
//
// 		// pack.Server.UseRestore()
// 		// pack.Server.UseApi()
//
// 		w.WriteHeader(http.StatusOK)
// 		return
// 	} else {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
//
// 	writeJson(w, http.StatusCreated, pack.InstanceService.GetLocal())
// 	// go pack.Server.Restart()
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
// //	@Tags		Servers
// //	@Produce	json
// //
// //	@Success	202
// //	@Failure	404
// //	@Failure	500
// //	@Router		/servers/reset [post]
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
// //	@Tags		Servers
// //	@Produce	json
// //
// //	@Param		serverId	path	string					true	"Server Id"
// //	@Param		request		body	rest.RestoreCoreParams	true	"Restore Params"
// //
// //	@Success	202
// //	@Failure	404
// //	@Failure	500
// //	@Router		/servers/{serverId}/restore [post]
// func restoreToCore(w http.ResponseWriter, r *http.Request) {
// 	restoreInfo, err := readCtxBody[rest.RestoreCoreParams](w, r)
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
