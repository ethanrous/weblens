package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func ping(ctx *gin.Context) {
	si := dataStore.GetServerInfo()
	if si == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "weblens not initialized"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": si.ServerId()})
}

func attachRemote(ctx *gin.Context) {
	si := dataStore.GetServerInfo()
	if si.ServerRole() == types.Backup {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "this weblens server is running in backup mode. core mode is required to attach a remote"})
		return
	}

	nr, err := readCtxBody[newServerBody](ctx)
	if err != nil {
		return
	}
	err = dataStore.NewRemote(nr.Id, nr.Name, types.WeblensApiKey(nr.UsingKey))
	if err != nil {
		if err == dataStore.ErrKeyInUse {
			ctx.Status(http.StatusConflict)
			return
		}
		switch err.(type) {
		case types.WeblensError:
			util.ShowErr(err)
		default:
			util.ErrTrace(err)
		}
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusCreated)
}

func getBackupSnapshot(ctx *gin.Context) {
	ts, err := strconv.Atoi(ctx.Query("since"))
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}
	since := time.UnixMilli(int64(ts))
	jes, err := dataStore.JournalSince(since)
	// jes[0].JournaledAt().String()
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	_, err = json.Marshal(gin.H{"journal": jes})
	util.ShowErr(err)

	ctx.JSON(http.StatusOK, gin.H{"journal": jes})
}

// /api/core/file/jfIjGtsl/content
// /api/core/file/jfIjGtsl/content

func getFileBytes(ctx *gin.Context) {
	fileId := types.FileId(ctx.Param("fileId"))
	f := dataStore.FsTreeGet(fileId)
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
	files := []map[string]any{}
	notFound := []types.FileId{}
	for _, id := range fIds {
		f := dataStore.FsTreeGet(id)
		if f == nil {
			notFound = append(notFound, id)
		} else {
			files = append(files, f.MarshalArchive())
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"files": files, "notFound": notFound})
}

func getRemotes(ctx *gin.Context) {
	srvs, err := dataStore.GetRemotes()
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"remotes": srvs})
}

func removeRemote(ctx *gin.Context) {
	body, err := readCtxBody[deleteRemoteBody](ctx)
	if err != nil {
		return
	}
	dataStore.DeleteRemote(body.RemoteId)

	ctx.Status(http.StatusOK)
}
