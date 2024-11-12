package http

import (
	"errors"
	"net/http"

	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/gin-gonic/gin"
)

// GetRemotes godoc
//
//	@ID			GetRemotes
//
//	@Summary	Get all remotes
//	@Tags		Remotes
//
//	@Security	ApiKeyAuth
//
//	@Success	200	{array}	rest.ServerInfo	"Server Info"
//	@Router		/remotes [get]
func getRemotes(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)

	remotes := pack.InstanceService.GetRemotes()
	localRole := pack.InstanceService.GetLocal().GetRole()

	var serverInfos []rest.ServerInfo
	for _, srv := range remotes {
		addr, _ := srv.GetAddress()
		client := pack.ClientService.GetClientByServerId(srv.ServerId())
		online := client != nil && client.Active.Load()

		var backupSize int64 = -1
		if localRole == models.BackupServerRole {
			backupSize = pack.FileService.Size(srv.ServerId())
		}
		serverInfos = append(serverInfos, rest.ServerInfo{
			Id:           srv.ServerId(),
			Name:         srv.GetName(),
			UsingKey:     srv.GetUsingKey(),
			Role:         srv.GetRole(),
			IsThisServer: srv.IsLocal(),
			Address:      addr,
			Online:       online,
			ReportedRole: srv.GetReportedRole(),
			LastBackup:   srv.LastBackup,
			BackupSize:   backupSize,
		})
	}

	writeJson(w, http.StatusOK, serverInfos)
}

// AttachRemote godoc
//
//	@ID			CreateRemote
//
//	@Summary	Create a new remote
//	@Tags		Remotes
//
//	@Security	ApiKeyAuth
//
//	@Param		request	body	rest.NewServerParams	true	"New Server Params"
//	@Success	201		{array}	rest.ServerInfo			"New Server Info"
//	@Router		/remotes [post]
func attachRemote(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	local := pack.InstanceService.GetLocal()
	if local.GetRole() == models.BackupServerRole {
		writeJson(w,
			http.StatusBadRequest,
			gin.H{"error": "this weblens server is running in backup mode. core mode is required to attach a remote"},
		)
		return
	}

	nr, err := readCtxBody[rest.NewServerParams](w, r)
	if err != nil {
		return
	}

	newRemote := models.NewInstance(nr.Id, nr.Name, nr.UsingKey, models.BackupServerRole, false, "", local.ServerId())

	err = pack.InstanceService.Add(newRemote)
	if err != nil {
		if errors.Is(err, werror.ErrKeyInUse) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		log.ErrTrace(err)
		writeJson(w, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = pack.AccessService.SetKeyUsedBy(nr.UsingKey, newRemote)
	if SafeErrorAndExit(err, w) {
		return
	}

	writeJson(w, http.StatusCreated, pack.InstanceService.GetLocal())
}

// DeleteRemote godoc
//
//	@ID			DeleteRemote
//
//	@Summary	Delete a remote
//	@Tags		Remotes
//
//	@Security	ApiKeyAuth
//
//	@Param		remoteId	query	string	true	"Server Id to delete"
//	@Success	200
//	@Router		/remotes [delete]
func removeRemote(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	remoteId := r.URL.Query().Get("remoteId")
	if remoteId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remote := pack.InstanceService.GetByInstanceId(remoteId)

	err := pack.InstanceService.Del(remote.DbId)
	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if key := remote.GetUsingKey(); key != "" {
		err = pack.AccessService.SetKeyUsedBy(key, nil)
		if SafeErrorAndExit(err, w) {
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
