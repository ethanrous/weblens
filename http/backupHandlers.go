package http

import (
	"net/http"
	"time"

	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/models"
	"github.com/gin-gonic/gin"
)

func launchBackup(ctx *gin.Context) {
	pack := getServices(ctx)

	if pack.InstanceService.GetLocal().IsCore() {
		serverId := ctx.Query("serverId")
		client := pack.ClientService.GetClientByServerId(serverId)
		msg := models.WsResponseInfo{
			EventTag: "do_backup",
		}
		err := client.Send(msg)
		if err != nil {
			log.ErrTrace(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		return
	}

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
