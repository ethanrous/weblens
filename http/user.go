package http

import (
	"net/http"

	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/gin-gonic/gin"
)

// Archive means sending ALL fields, including password and token information
func getUsersArchive(ctx *gin.Context) {
	pack := getServices(ctx)
	usersIter, err := pack.UserService.GetAll()
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	var users []map[string]any
	for u := range usersIter {
		archive, err := u.FormatArchive()
		if err != nil {
			log.ErrTrace(err)
		}
		users = append(users, archive)
	}

	ctx.JSON(http.StatusOK, users)
}

func createUser(ctx *gin.Context) {
	pack := getServices(ctx)
	userInfo, err := readCtxBody[newUserBody](ctx)
	if err != nil {
		return
	}

	u, err := models.NewUser(userInfo.Username, userInfo.Password, userInfo.Admin, userInfo.AutoActivate)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	err = pack.FileService.CreateUserHome(u)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	err = pack.UserService.Add(u)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	ctx.Status(http.StatusCreated)
}
