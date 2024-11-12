package http

import (
	"net/http"

	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/models"
	"github.com/go-chi/chi/v5"
)

// GetFileShare godoc
//
//	@ID			GetFileShare
//
//	@Summary	Get a file share
//	@Tags		Share
//	@Produce	json
//	@Param		shareId	path		string				true	"Share Id"
//	@Success	200		{object}	models.FileShare	"File Share"
//	@Failure	404
//	@Router		/share/{shareId} [get]
func getFileShare(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(w, r)
	if SafeErrorAndExit(err, w) {
		return
	}
	shareId := models.ShareId(chi.URLParam(r, "shareId"))

	share := pack.ShareService.Get(shareId)
	if share == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fileShare, ok := share.(*models.FileShare)
	if !ok {
		log.Warning.Printf(
			"%s tried to get share [%s] as a fileShare (is %s)", u.GetUsername(), shareId, share.GetShareType(),
		)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	writeJson(w, http.StatusOK, fileShare)
}
