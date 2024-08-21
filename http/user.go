package http

import (
	"errors"
	"net/http"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/gin-gonic/gin"
)

// Archive means sending ALL fields, including password and token information
func getUsersArchive(ctx *gin.Context) {
	us, err := UserService.GetAll()
	if err != nil {
		wlog.ShowErr(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	archive := internal.Map(
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

	u, err := weblens.NewUser(userInfo.Username, userInfo.Password, userInfo.Admin, userInfo.AutoActivate)
	if err != nil {
		if errors.Is(err, types.ErrUserAlreadyExists) {
			ctx.Status(http.StatusConflict)
			return
		}
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	// homeDir, err := u.CreateHomeFolder()
	// if err != nil {
	// 	return nil, err
	// }
	//
	// newUser.homeFolder = homeDir
	// newUser.trashFolder = homeDir.GetChildren()[0]

	err = UserService.Add(u)
	if err != nil {
		wlog.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusCreated)
}
