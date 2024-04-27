package main

import (
	"os"
	"time"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/routes"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const retries = 10

func main() {
	sw := util.NewStopwatch("Initialization")
	godotenv.Load()

	if util.IsDevMode() {
		util.Debug.Println("Initializing weblens in development mode")
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	sw.Lap()

	rq := routes.NewRequester()
	srvInfo := dataStore.GetServerInfo()
	if srvInfo.ServerRole() == types.Backup {
		checkCoreExists(rq, sw)
	}

	routes.VerifyClientManager()
	tt := dataProcess.VerifyTaskTracker()
	dataStore.SetTasker(dataProcess.NewTaskPool(false, nil))

	routes.UploadTasker = dataProcess.NewTaskPool(false, nil)
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

	store := dataStore.NewStore(rq)

	util.Info.Println("Loading users...")
	err = store.LoadUsers()
	util.FailOnError(err, "Failed to load users")
	sw.Lap("Users init")

	// 								100MB
	et := dataProcess.InitGExif(1000 * 1000 * 100)
	if et == nil {
		panic("Exiftool is nil")
	}
	dataStore.SetExiftool(et)
	sw.Lap("Init global exiftool")

	// Enable the worker pool held by the task tracker
	// loading the filesystem might dispatch tasks,
	// so we have to start the pool first
	tt.StartWP()
	sw.Lap("Global worker pool enabled")

	// Load filesystem
	util.Info.Println("Loading filesystem...")
	dataStore.FsInit()
	sw.Lap("FS init")
	util.Info.Println("Initialized Filesystem")

	dataStore.MediaInit()
	sw.Lap("Media init")
	util.Info.Println("Initialized Media Map")

	dataStore.VerifyAlbumsMedia()
	sw.Lap("Albums init")

	dataStore.LoadAllShares()
	sw.Lap("Shares init")

	dataStore.InitApiKeyMap()
	sw.Lap("Api key map init")

	// The global broadcaster is disabled by default so all of the
	// initial loading of the filesystem (that was just done above) doesn't
	// try to broadcast for every file that exists. So it must be enabled here
	routes.Caster.DropBuffer()
	sw.Lap("Global caster enabled")

	if srvInfo.ServerRole() == types.Backup {
		go dataProcess.BackupD(time.Minute, rq)
		sw.Lap("Init backup sleeper")
	}

	sw.Stop()
	sw.PrintResults(false)

	util.Info.Printf("Weblens loaded. %d files and %d medias\n", dataStore.GetTreeSize(), dataStore.GetMediaMapSize())

	for {
		routes.DoRoutes()
	}

}

func checkCoreExists(rq types.Requester, sw util.Stopwatch) {
	connected := false
	i := 0
	for i = range retries {
		if rq.PingCore() {
			connected = true
			break
		}
		time.Sleep(time.Millisecond * 500)
	}
	if !connected {
		util.Error.Println("Failed to ping core server")
		os.Exit(1)
	}
	sw.Lap("Connected to core server after ", i, " retries")
}
