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
func getRemotes(ctx *gin.Context) {
	pack := getServices(ctx)

	remotes := pack.InstanceService.GetRemotes()
	localRole := pack.InstanceService.GetLocal().GetRole()

	var serverInfos []rest.ServerInfo
	for _, srv := range remotes {
		addr, _ := srv.GetAddress()
		client := pack.ClientService.GetClientByServerId(srv.ServerId())
		online := client != nil && client.Active.Load()

		var backupSize int64 = -1
		if localRole == models.BackupServer {
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

	ctx.JSON(http.StatusOK, serverInfos)
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
func attachRemote(ctx *gin.Context) {
	pack := getServices(ctx)
	local := pack.InstanceService.GetLocal()
	if local.GetRole() == models.BackupServer {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{"error": "this weblens server is running in backup mode. core mode is required to attach a remote"},
		)
		return
	}

	nr, err := readCtxBody[rest.NewServerParams](ctx)
	if err != nil {
		return
	}

	newRemote := models.NewInstance(nr.Id, nr.Name, nr.UsingKey, models.BackupServer, false, "", local.ServerId())

	err = pack.InstanceService.Add(newRemote)
	if err != nil {
		if errors.Is(err, werror.ErrKeyInUse) {
			ctx.Status(http.StatusConflict)
			return
		}

		log.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = pack.AccessService.SetKeyUsedBy(nr.UsingKey, newRemote)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ctx.JSON(http.StatusCreated, pack.InstanceService.GetLocal())
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
func removeRemote(ctx *gin.Context) {
	pack := getServices(ctx)
	remoteId := ctx.Query("remoteId")
	if remoteId == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}

	remote := pack.InstanceService.GetByInstanceId(remoteId)

	err := pack.InstanceService.Del(remote.DbId)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	if key := remote.GetUsingKey(); key != "" {
		err = pack.AccessService.SetKeyUsedBy(key, nil)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
	}

	ctx.Status(http.StatusOK)
}
