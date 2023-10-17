package main

import (
	"os"

	"github.com/ethrousseau/weblens/api/routes"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	os.Stat(util.GetMediaRoot())

	router := gin.Default()

	routes.AddApiRoutes(router)
	if !util.IsDevMode() {
		routes.AddUiRoutes(router)
	}

	router.Run("0.0.0.0:8080")
}