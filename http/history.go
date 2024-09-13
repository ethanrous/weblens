package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/gin-gonic/gin"
)

func getLifetimesSince(ctx *gin.Context) {
	pack := getServices(ctx)
	log.Debug.Println(ctx.Params)
	millisString := ctx.Param("timestamp")
	if millisString == "" {
		log.Error.Println("No timestamp given trying to get lifetimes since date")
		ctx.Status(http.StatusBadRequest)
		return
	}

	millis, err := strconv.ParseInt(millisString, 10, 64)
	if err != nil || millis < 0 {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	date := time.UnixMilli(millis)

	lifetimes, err := pack.FileService.GetMediaJournal().GetLifetimesSince(date)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.JSON(http.StatusOK, lifetimes)
}

func getHistory(ctx *gin.Context) {
	pack := getServices(ctx)
	lts := pack.FileService.GetMediaJournal().GetAllLifetimes()
	ctx.JSON(http.StatusOK, lts)
}
