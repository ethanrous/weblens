package routes

import (
	"net/http"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/gin-gonic/gin"
)

func launchBackup(ctx *gin.Context) {
	rs := types.SERV.InstanceService.GetRemotes()

	t := types.SERV.TaskDispatcher.Backup(rs[0].ServerId(), types.SERV.Caster)

	// if _, stat := t.Status(); stat != dataProcess.TaskSuccess {
	// 	ctx.Status(http.StatusInternalServerError)
	// 	return
	// }
	user := getUserFromCtx(ctx)
	wsClient := types.SERV.ClientManager.GetClientByUsername(user.GetUsername())

	time.Sleep(time.Millisecond * 10)
	exited, _ := t.Status()
	for t.GetChildTaskPool() == nil && !exited {
		time.Sleep(time.Millisecond * 100)
	}

	acc := dataStore.NewAccessMeta(user)
	wsClient.Subscribe(types.SubId(t.GetChildTaskPool().ID()), types.PoolSubscribe, acc)

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
