package routes

import (
	"errors"
	"net/http"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/dataStore/instance"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func ping(ctx *gin.Context) {
	local := types.SERV.InstanceService.GetLocal()
	if local == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "weblens not initialized"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": local.ServerId()})
}

func attachRemote(ctx *gin.Context) {
	local := types.SERV.InstanceService.GetLocal()
	if local.ServerRole() == types.Backup {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "this weblens server is running in backup mode. core mode is required to attach a remote"})
		return
	}

	nr, err := readCtxBody[newServerBody](ctx)
	if err != nil {
		return
	}

	newRemote := instance.New(nr.Id, nr.Name, nr.UsingKey, types.Backup, false, "")

	err = types.SERV.InstanceService.Add(newRemote)
	if err != nil {
		if errors.Is(err, dataStore.ErrKeyInUse) {
			ctx.Status(http.StatusConflict)
			return
		}

		util.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusCreated)
}

func getBackupSnapshot(ctx *gin.Context) {
	// ts, err := strconv.Atoi(ctx.Query("since"))
	// if err != nil {
	// 	util.ShowErr(err)
	// 	ctx.Status(http.StatusBadRequest)
	// 	return
	// }
	// since := time.UnixMilli(int64(ts))
	// jes, err := dataStore.JournalSince(since)
	// // jes[0].JournaledAt().String()
	// if err != nil {
	// 	util.ShowErr(err)
	// 	ctx.Status(http.StatusInternalServerError)
	// 	return
	// }
	//
	// _, err = json.Marshal(gin.H{"journal": jes})
	// util.ShowErr(err)
	//
	// ctx.JSON(http.StatusOK, gin.H{"journal": jes})
	ctx.Status(http.StatusNotImplemented)
}

func getFileBytes(ctx *gin.Context) {
	fileId := types.FileId(ctx.Param("fileId"))
	f := types.SERV.FileTree.Get(fileId)
	if f == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	if f.IsDir() {
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.File(f.GetAbsPath())
}

func getFilesMeta(ctx *gin.Context) {
	fIds, err := readCtxBody[[]types.FileId](ctx)
	if err != nil {
		return
	}
	var files []map[string]any
	var notFound []types.FileId
	for _, id := range fIds {
		f := types.SERV.FileTree.Get(id)
		if f == nil {
			notFound = append(notFound, id)
		} else {
			files = append(files, f.MarshalArchive())
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"files": files, "notFound": notFound})
}

func getRemotes(ctx *gin.Context) {
	srvs := types.SERV.InstanceService.GetRemotes()
	ctx.JSON(http.StatusOK, gin.H{"remotes": srvs})
}

func removeRemote(ctx *gin.Context) {
	body, err := readCtxBody[deleteRemoteBody](ctx)
	if err != nil {
		return
	}

	err = types.SERV.InstanceService.Del(body.RemoteId)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}
