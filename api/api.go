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
	sw := util.NewStopwatch("Initialization")
	godotenv.Load()

	if util.IsDevMode() {
		util.Debug.Println("Initializing weblens in development mode")
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	sw.Lap()
	routes.VerifyClientManager()
	tt := dataProcess.VerifyTaskTracker()
	dataStore.SetTasker(dataProcess.NewWorkQueue())

	routes.UploadTasker = dataProcess.NewWorkQueue()
	routes.UploadTasker.MarkGlobal()

	sw.Lap("Verify cm, tt and set routes and ds queue")

	dataProcess.SetCaster(routes.Caster)
	dataStore.SetCaster(routes.Caster)
	dataStore.SetVoidCaster(routes.VoidCaster)
	sw.Lap("Set casters")

	err := dataStore.ClearTempDir()
	util.FailOnError(err, "Failed to clear temporary directory on startup")
	sw.Lap("Clear tmp dir")

	err = dataStore.ClearTakeoutDir()
	util.FailOnError(err, "Failed to clear takeout directory on startup")
	sw.Lap("Clear takeout dir")

	err = dataStore.InitMediaTypeMaps()
	util.FailOnError(err, "Failed to initialize media type map")
	sw.Lap("Init type map")

	// // Load filesystem
	// util.Info.Println("Loading filesystem...")
	// dataStore.FsInit()
	// sw.Lap("FS init")
	// util.Info.Println("Initialized Filesystem")

	dataStore.MediaInit()
	sw.Lap("Media init")
	util.Info.Println("Initialized Media Map")

	dataStore.LoadAllShares()
	sw.Lap("Shares init")

	// The global broadcaster is disbled by default so all of the
	// initial loading of the filesystem (that was just done above) doesn't
	// try to broadcast for every file that exists. So it must be enabled here
	routes.Caster.DropBuffer()
	routes.Caster.AutoflushEnable()
	sw.Lap("Global caster enabled")

	// 								100MB
	et := dataProcess.InitGExif(1000 * 1000 * 100)
	if et == nil {
		panic("Exiftool is nil")
	}
	dataStore.SetExiftool(et)
	sw.Lap("Init global exiftool")

	// Enable the worker pool heald by the task tracker
	tt.StartWP()
	sw.Lap("Global worker pool enabled")

	router := gin.Default()

	routes.AddApiRoutes(router)
	if !util.DetachUi() {
		routes.AddUiRoutes(router)
	}
	if util.IsDevMode() {
		routes.AttachProfiler(router)
	}
	sw.Lap("Gin routes added")
	sw.Stop()

	sw.PrintResults()

	util.Info.Println("Weblens loaded. Starting router...")

	router.Run(util.GetRouterIp() + ":" + util.GetRouterPort())
}
