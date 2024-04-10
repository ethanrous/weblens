package routes

import (
	"net/http"

	"github.com/ethrousseau/weblens/api/dataProcess"
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
	if si.ServerRole() == types.BackupMode {
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
	rq := NewRequester()
	t := dataProcess.GetGlobalQueue().Backup(rq)
	t.Wait()

	_, status := t.Status()
	if status == dataProcess.TaskSuccess {
		res := t.GetResult()
		ctx.JSON(http.StatusOK, res)
	} else {
		ctx.Status(http.StatusInternalServerError)
		return
	}

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
