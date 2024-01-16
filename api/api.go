package main

import (
	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/routes"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	dataProcess.SetCaster(routes.Caster)
	dataStore.SetCaster(routes.Caster)
	dataStore.SetTasker(dataProcess.NewWorkQueue())

	err := dataStore.ClearTempDir()
	util.FailOnError(err, "Failed to clear temporary directory on startup")

	err = dataStore.ClearTakeoutDir()
	util.FailOnError(err, "Failed to clear takeout directory on startup")

	err = dataStore.InitMediaTypeMaps()
	util.FailOnError(err, "Failed to initialize media type map")

	// Load filesystem
	dataStore.FsInit()

	// The global broadcaster is disbled by default so all of the
	// initial loading of the filesystem (that was just done above) doesn't
	// try to broadcast for every file that exists. So it must be enabled here
	routes.Caster.Enable()

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
