package http

import (
	"net/http"
)

func clearCache(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	return
	// pack := getServices(r)
	// err := pack.MediaService.NukeCache()
	// if err != nil {
	// 	log.ShowErr(err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	//
	// w.WriteHeader(http.StatusOK)

}

func resetServer(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	err := pack.InstanceService.ResetAll()
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)

	// pack.Server.Restart()
}
