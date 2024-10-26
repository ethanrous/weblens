package http

import (
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/gin-gonic/gin"
)

func getLifetimesSince(ctx *gin.Context) {
	pack := getServices(ctx)

	millisString := ctx.Query("timestamp")
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

	lifetimes, err := pack.FileService.GetJournalByTree("USERS").GetLifetimesSince(date)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	ctx.JSON(http.StatusOK, lifetimes)
}

func doFullBackup(ctx *gin.Context) {
	pack := getServices(ctx)
	instance := getInstanceFromCtx(ctx)

	millisString := ctx.Query("timestamp")
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

	since := time.UnixMilli(millis)
	usersJournal := pack.FileService.GetJournalByTree("USERS")
	lts, err := usersJournal.GetLifetimesSince(since)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}
	usersItter, err := pack.UserService.GetAll()
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}
	users := slices.Collect(usersItter)

	instances := pack.InstanceService.GetRemotes()
	instances = append(instances, pack.InstanceService.GetLocal())

	usingKey, err := pack.AccessService.GetApiKey(instance.GetUsingKey())
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	owner := pack.UserService.Get(usingKey.Owner)

	keys, err := pack.AccessService.GetAllKeys(owner)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	res := models.BackupBody{
		FileHistory:    lts,
		LifetimesCount: len(usersJournal.GetAllLifetimes()),
		Users:          users,
		Instances:      instances,
		ApikKeys:       keys,
	}
	ctx.JSON(http.StatusOK, res)
}
