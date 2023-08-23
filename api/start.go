package api

import (
	"github.com/gin-gonic/gin"

	"github.com/EthRousseau/weblens/ui"
)

func Start() {
	router := gin.Default()

	addRoutes(router)
	ui.AddRoutes(router)

	router.Run("127.0.0.1:4000")
}
