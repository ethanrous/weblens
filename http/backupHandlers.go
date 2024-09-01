package http

import (
	"net/http"
	"time"

	"github.com/ethrousseau/weblens/models"
	"github.com/gin-gonic/gin"
)

func launchBackup(ctx *gin.Context) {
	pack := getServices(ctx)
	core := pack.InstanceService.GetCore()

	backupMeta := models.BackupMeta{
		RemoteId:        core.ServerId(),
		InstanceService: pack.InstanceService,
	}
	t, err := pack.TaskService.DispatchJob(models.BackupTask, backupMeta, nil)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	user := getUserFromCtx(ctx)
	wsClient := pack.ClientService.GetClientByUsername(user.GetUsername())

	// exited, _ := t.Status()
	// for t.GetChildTaskPool() == nil && !exited {
	// 	time.Sleep(time.Millisecond * 100)
	// }

	_, _, err = pack.ClientService.Subscribe(
		wsClient, models.SubId(t.TaskId()), models.TaskSubscribe, time.Now(), nil,
	)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func getSnapshots(ctx *gin.Context) {
	// jes, err := dataStore.GetSnapshots()
	// if err != nil {
	// 	util.ShowErr(err)
	// 	ctx.Status(comm.StatusInternalServerError)
	// 	return
	// }
	//
	// ctx.JSON(comm.StatusOK, gin.H{"snapshots": jes})
	ctx.Status(http.StatusNotImplemented)
}
