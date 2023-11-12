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

	var ip string

	routes.AddApiRoutes(router)
	if !util.IsDevMode() {
		ip = "0.0.0.0"
		routes.AddUiRoutes(router)
	} else {
		ip = "127.0.0.1"
	}

	router.Run(ip + ":8080")
}