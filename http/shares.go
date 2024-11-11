package http

import (
	"net/http"

	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/models"
	"github.com/gin-gonic/gin"
)

// GetFileShare godoc
//	@ID			GetFileShare
//
//	@Summary	Get a file share
//	@Tags		Share
//	@Produce	json
//	@Param		shareId	path		string				true	"Share Id"
//	@Success	200		{object}	models.FileShare	"File Share"
//	@Failure	404
//	@Router		/share/{shareId} [get]
func getFileShare(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	shareId := models.ShareId(ctx.Param("shareId"))

	share := pack.ShareService.Get(shareId)
	if share == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	fileShare, ok := share.(*models.FileShare)
	if !ok {
		log.Warning.Printf(
			"%s tried to get share [%s] as a fileShare (is %s)", u.GetUsername(), shareId, share.GetShareType(),
		)
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, fileShare)
}
