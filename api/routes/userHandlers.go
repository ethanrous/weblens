package routes

import (
	"net/http"

	"github.com/ethrousseau/weblens/api/types"
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
