package http

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/jobs"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/ethanrous/weblens/service/mock"
)

func attachNewCoreRemote(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)

	body, err := readCtxBody[rest.NewCoreBody](w, r)
	if err != nil {
		return
	}

	newCore, err := pack.InstanceService.AttachRemoteCore(body.CoreAddress, body.UsingKey)
	if SafeErrorAndExit(err, w) {
		return
	}

	mockJournal := mock.NewHollowJournalService()
	newTree, err := fileTree.NewFileTree(filepath.Join(env.GetDataRoot(), newCore.ServerId()), newCore.ServerId(), mockJournal, false)
	if err != nil {
		log.ErrTrace(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pack.FileService.AddTree(newTree)

	err = WebsocketToCore(newCore, pack)
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

func launchBackup(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)

	serverId := r.URL.Query().Get("serverId")
	if serverId == "" {
		w.WriteHeader(http.StatusBadRequest)
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		core := pack.InstanceService.GetByInstanceId(serverId)
		if core == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		t, err := jobs.BackupOne(core, pack)
		if err != nil {
			log.ErrTrace(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		u, err := getUserFromCtx(w, r)
		if SafeErrorAndExit(err, w) {
			return
		}
		log.Debug.Printf("User: %s", u.GetUsername())
		wsClient := pack.ClientService.GetClientByUsername(u.GetUsername())

		_, _, err = pack.ClientService.Subscribe(
			wsClient, t.TaskId(), models.TaskSubscribe, time.Now(), nil,
		)
		if err != nil {
			log.ErrTrace(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func restoreToCore(w http.ResponseWriter, r *http.Request) {
	restoreInfo, err := readCtxBody[rest.RestoreCoreBody](w, r)

	if err != nil {
		return
	}

	pack := getServices(r)

	core := pack.InstanceService.GetByInstanceId(restoreInfo.ServerId)
	if core == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = core.SetAddress(restoreInfo.HostUrl)
	if SafeErrorAndExit(err, w) {
		return
	}

	meta := models.RestoreCoreMeta{
		Local: pack.InstanceService.GetLocal(),
		Core:  core,
		Pack:  pack,
	}

	_, err = pack.TaskService.DispatchJob(models.RestoreCoreTask, meta, nil)
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
