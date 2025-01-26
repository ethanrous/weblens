package http

import (
	"net/http"
	"time"

	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/jobs"
	"github.com/ethanrous/weblens/models"
	"github.com/go-chi/chi/v5"
)

// LaunchBackup godoc
//
//	@ID			LaunchBackup
//
//	@Summary	Launch backup on a server
//	@Tags		Servers
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Param		serverId	path	string	true	"Server ID"
//
//	@Success	200
//	@Router		/servers/{serverId}/backup [post]
func launchBackup(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)

	serverId := chi.URLParam(r, "serverId")
	if serverId == "" {
		SafeErrorAndExit(werror.ErrNoServerId, w)
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

		u, err := getUserFromCtx(r, true)
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
