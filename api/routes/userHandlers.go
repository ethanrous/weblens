package routes

import (
	"net/http"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/gin-gonic/gin"
)

// Archive means sending ALL fields, including password and token information
func getUsersArchive(ctx *gin.Context) {
	rq := NewRequester()
	store := dataStore.NewStore(rq)

	users := store.GetUsers()
	arc := dataStore.UserArray(users).MarshalArchive()
	ctx.JSON(http.StatusOK, arc)
}
