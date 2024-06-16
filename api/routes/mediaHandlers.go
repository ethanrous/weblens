package routes

import (
	"net/http"

	"github.com/ethrousseau/weblens/api/dataStore/media"
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
		m := media.MediaMapGet(mId)
		if m == nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		medias[i] = m

	}

	err = media.HideMedia(medias)
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

	anchor := media.MediaMapGet(body.AnchorId)
	if anchor == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	extras := util.Map(body.MediaIds, func(mId types.ContentId) types.Media { return media.MediaMapGet(mId) })

	err = media.AdjustMediaDates(anchor, body.NewTime, extras)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.Status(http.StatusOK)
}
