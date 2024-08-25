package comm

import (
	"errors"
	"net/http"

	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/models"
	"github.com/gin-gonic/gin"
)

// Archive means sending ALL fields, including password and token information
func getUsersArchive(ctx *gin.Context) {
	users := UserService.GetAll()

	var archives []map[string]any
	for u := range users {
		archive, err := u.FormatArchive()
		if err != nil {
			log.ErrTrace(err)
		}
		archives = append(archives, archive)
	}

	ctx.JSON(http.StatusOK, archives)
}

func createUser(ctx *gin.Context) {
	userInfo, err := readCtxBody[newUserBody](ctx)
	if err != nil {
		return
	}

	u, err := models.NewUser(userInfo.Username, userInfo.Password, userInfo.Admin, userInfo.AutoActivate)
	if err != nil {
		if errors.Is(err, werror.ErrUserAlreadyExists) {
			ctx.Status(http.StatusConflict)
			return
		}
		log.ShowErr(err)
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
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusCreated)
}
