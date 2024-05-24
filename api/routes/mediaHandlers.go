package routes

import (
	"net/http"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func hideMedia(ctx *gin.Context) {
	body, err := readCtxBody[mediaIdsBody](ctx)
	if err != nil {
		return
	}

	medias := make([]types.Media, len(body.MediaIds))
	for i, mId := range body.MediaIds {
		m := dataStore.MediaMapGet(mId)
		if m == nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		medias[i] = m

	}

	err = dataStore.HideMedia(medias)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func adjustMediaDate(ctx *gin.Context) {
	body, err := readCtxBody[mediaTimeBody](ctx)
	if err != nil {
		return
	}

	anchor := dataStore.MediaMapGet(body.AnchorId)
	if anchor == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	extras := util.Map(body.MediaIds, func(mId types.ContentId) types.Media { return dataStore.MediaMapGet(mId) })

	err = dataStore.AdjustMediaDates(anchor, body.NewTime, extras)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.Status(http.StatusOK)
}
