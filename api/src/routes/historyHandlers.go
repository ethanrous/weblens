package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func getLifetimesSince(ctx *gin.Context) {
	millisString := ctx.Param("timestamp")
	millis, err := strconv.ParseInt(millisString, 10, 64)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	date := time.UnixMilli(millis)

	lifetimes, err := types.SERV.StoreService.GetLifetimesSince(date)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, lifetimes)
}

func getHistory(ctx *gin.Context) {
	lts := types.SERV.FileTree.GetJournal().GetAllLifetimes()
	ctx.JSON(http.StatusOK, lts)
}
