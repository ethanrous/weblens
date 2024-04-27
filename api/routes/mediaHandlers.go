package routes

import (
	"net/http"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func hideMedia(ctx *gin.Context) {
	mediaId := types.ContentId(ctx.Param("mediaId"))
	m := dataStore.MediaMapGet(mediaId)
	if m == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	err := dataStore.HideMedia(m)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}
