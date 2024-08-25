package comm

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ethrousseau/weblens/internal/log"
	"github.com/gin-gonic/gin"
)

func getLifetimesSince(ctx *gin.Context) {
	millisString := ctx.Param("timestamp")
	millis, err := strconv.ParseInt(millisString, 10, 64)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	date := time.UnixMilli(millis)

	lifetimes, err := FileService.GetMediaJournal().GetLifetimesSince(date)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, lifetimes)
}

func getHistory(ctx *gin.Context) {
	lts := FileService.GetMediaJournal().GetAllLifetimes()
	ctx.JSON(http.StatusOK, lts)
}
