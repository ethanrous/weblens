package comm

import (
	"net/http"
	"time"

	"github.com/ethrousseau/weblens/models"
	"github.com/gin-gonic/gin"
)

func launchBackup(ctx *gin.Context) {
	core := InstanceService.GetCore()

	backupMeta := models.BackupMeta{
		RemoteId:        core.ServerId(),
		InstanceService: InstanceService,
	}
	t, err := TaskService.DispatchJob(models.BackupTask, backupMeta, nil)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	user := getUserFromCtx(ctx)
	wsClient := ClientService.GetClientByUsername(user.GetUsername())

	// exited, _ := t.Status()
	// for t.GetChildTaskPool() == nil && !exited {
	// 	time.Sleep(time.Millisecond * 100)
	// }

	_, _, err = ClientService.Subscribe(
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
