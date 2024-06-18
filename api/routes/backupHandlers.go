package routes

import (
	"net/http"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/gin-gonic/gin"
)

func launchBackup(ctx *gin.Context) {
	rs := types.SERV.InstanceService.GetRemotes()

	t := types.SERV.TaskDispatcher.Backup(rs[0].ServerId(), types.SERV.Requester, types.SERV.FileTree)
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
