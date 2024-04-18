package routes

import (
	"net/http"
	"slices"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/gin-gonic/gin"
)

func publicGetUsers(ctx *gin.Context) {
	users := Store.GetUsers()
	if slices.ContainsFunc(users, func(u types.User) bool { return u.IsOwner() }) && dataStore.GetServerInfo() != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"users": users})

}
