package http

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/jobs"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/service/mock"
	"github.com/gin-gonic/gin"
)

func attachNewCoreRemote(ctx *gin.Context) {
	pack := getServices(ctx)

	body, err := readCtxBody[newCoreBody](ctx)
	if err != nil {
		return
	}

	newCore, err := pack.InstanceService.AttachRemoteCore(body.CoreAddress, body.UsingKey)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	mockJournal := mock.NewHollowJournalService()
	newTree, err := fileTree.NewFileTree(filepath.Join(env.GetDataRoot(), newCore.ServerId()), newCore.ServerId(), mockJournal, false)
	if err != nil {
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	pack.FileService.AddTree(newTree)

	err = WebsocketToCore(newCore, pack)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.Status(http.StatusOK)
}

func launchBackup(ctx *gin.Context) {
	pack := getServices(ctx)

	serverId := ctx.Query("serverId")
	if serverId == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}

	local := pack.InstanceService.GetLocal()

	// If the local is core, we send the backup request to the specied backup server
	if local.IsCore() {
		client := pack.ClientService.GetClientByServerId(serverId)
		msg := models.WsResponseInfo{
			EventTag: "do_backup",
			Content:  models.WsC{"coreId": local.ServerId()},
		}
		err := client.Send(msg)
		if err != nil {
			log.ErrTrace(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
	} else {
		core := pack.InstanceService.GetByInstanceId(serverId)
		if core == nil {
			ctx.Status(http.StatusNotFound)
			return
		}

		t, err := jobs.BackupOne(core, pack)
		if err != nil {
			log.ErrTrace(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		user := getUserFromCtx(ctx)
		log.Debug.Printf("User: %s", user.GetUsername())
		wsClient := pack.ClientService.GetClientByUsername(user.GetUsername())

		_, _, err = pack.ClientService.Subscribe(
			wsClient, t.TaskId(), models.TaskSubscribe, time.Now(), nil,
		)
		if err != nil {
			log.ErrTrace(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
	}

	ctx.Status(http.StatusOK)
}

func restoreToCore(ctx *gin.Context) {
	restoreInfo, err := readCtxBody[restoreCoreBody](ctx)

	if err != nil {
		return
	}

	pack := getServices(ctx)

	core := pack.InstanceService.GetByInstanceId(restoreInfo.ServerId)
	if core == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	err = core.SetAddress(restoreInfo.HostUrl)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	meta := models.RestoreCoreMeta{
		Local: pack.InstanceService.GetLocal(),
		Core:  core,
		Pack:  pack,
	}

	_, err = pack.TaskService.DispatchJob(models.RestoreCoreTask, meta, nil)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.Status(http.StatusAccepted)
}
