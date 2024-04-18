package routes

import (
	"net/http"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func launchBackup(ctx *gin.Context) {
	rq := NewRequester()

	rs, err := dataStore.GetRemotes()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	t := dataProcess.GetGlobalQueue().Backup(rs[0].ServerId(), rq)
	t.Wait()
	if _, stat := t.Status(); stat != dataProcess.TaskSuccess {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

func getSnapshots(ctx *gin.Context) {
	jes, err := dataStore.GetSnapshots()
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"snapshots": jes})
}
