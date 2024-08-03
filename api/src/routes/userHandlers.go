package routes

import (
	"errors"
	"net/http"

	"github.com/ethrousseau/weblens/api/dataStore/user"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/ethrousseau/weblens/api/util/wlog"
	"github.com/gin-gonic/gin"
)

// Archive means sending ALL fields, including password and token information
func getUsersArchive(ctx *gin.Context) {
	us, err := types.SERV.UserService.GetAll()
	if err != nil {
		wlog.ShowErr(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	archive := util.Map(
		us, func(u types.User) map[string]any {
			ar, err := u.FormatArchive()
			if err != nil {
				wlog.ShowErr(err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return nil
			}
			return ar
		},
	)

	ctx.JSON(http.StatusOK, archive)
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
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	err = types.SERV.UserService.Add(u)
	if err != nil {
		wlog.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusCreated)
}
