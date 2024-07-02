package routes

import (
	"errors"
	"net/http"

	"github.com/ethrousseau/weblens/api/dataStore/user"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

// Archive means sending ALL fields, including password and token information
func getUsersArchive(ctx *gin.Context) {
	us, err := types.SERV.Requester.GetCoreUsers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, us)
}

func createUser(ctx *gin.Context) {
	userInfo, err := readCtxBody[newUserBody](ctx)
	if err != nil {
		return
	}

	u, err := user.New(userInfo.Username, userInfo.Password, userInfo.Admin, userInfo.AutoActivate)
	if err != nil {
		if errors.Is(err, types.ErrUserAlreadyExists) {
			ctx.Status(http.StatusConflict)
			return
		}
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	err = types.SERV.UserService.Add(u)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusCreated)
}
