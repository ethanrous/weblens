package routes

import (
	"net/http"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/gin-gonic/gin"
)

func launchBackup(ctx *gin.Context) {
	rs, err := dataStore.GetRemotes()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	t := rc.TaskDispatcher.Backup(rs[0].ServerId(), rc.Requester, rc.FileTree)
	t.Wait()
	if _, stat := t.Status(); stat != dataProcess.TaskSuccess {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

func getSnapshots(ctx *gin.Context) {
	// jes, err := dataStore.GetSnapshots()
	// if err != nil {
	// 	util.ShowErr(err)
	// 	ctx.Status(http.StatusInternalServerError)
	// 	return
	// }
	//
	// ctx.JSON(http.StatusOK, gin.H{"snapshots": jes})
	ctx.Status(http.StatusNotImplemented)
}
