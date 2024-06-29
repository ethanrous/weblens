package main

import (
	"os"
	"runtime"
	"time"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/dataStore/album"
	"github.com/ethrousseau/weblens/api/dataStore/database"
	"github.com/ethrousseau/weblens/api/dataStore/filetree"
	"github.com/ethrousseau/weblens/api/dataStore/history"
	"github.com/ethrousseau/weblens/api/dataStore/instance"
	"github.com/ethrousseau/weblens/api/dataStore/media"
	"github.com/ethrousseau/weblens/api/dataStore/share"
	"github.com/ethrousseau/weblens/api/dataStore/user"
	"github.com/ethrousseau/weblens/api/routes"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const retries = 10

func main() {
	sw := util.NewStopwatch("Initialization")
	err := godotenv.Load()
	if err != nil {
		util.ShowErr(err)
	}

	if util.IsDevMode() {
		util.Debug.Println("Initializing weblens in development mode")
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	sw.Lap()

	// Gather global services

	dbService := database.New(util.GetMongoURI(), util.GetMongoDBName())
	types.SERV.SetDatabase(dbService)

	instanceService := instance.NewService()
	err = instanceService.Init(dbService)
	if err != nil {
		panic(err)
	}
	types.SERV.SetInstance(instanceService)

	ft := filetree.NewFileTree()
	err = ft.Init(dbService)
	if err != nil {
		panic(err)
	}
	types.SERV.SetFileTree(ft)

	userService := user.NewService()
	err = userService.Init(dbService)
	if err != nil {
		panic(err)
	}
	types.SERV.SetUserService(userService)

	instanceService.GetLocal().SetUserCount(types.SERV.UserService.Size())

	mediaTypeServ := media.NewTypeService()
	mediaService := media.NewRepo(mediaTypeServ)
	err = mediaService.Init(dbService)
	if err != nil {
		panic(err)
	}
	types.SERV.SetMediaRepo(mediaService)

	albumService := album.NewService()
	err = albumService.Init(dbService)
	if err != nil {
		panic(err)
	}
	types.SERV.SetAlbumService(albumService)

	clientService := routes.NewClientManager()
	types.SERV.SetClientService(clientService)

	workerPool, taskDispatcher := dataProcess.NewWorkerPool(runtime.NumCPU() - 2)
	types.SERV.SetWorkerPool(workerPool)
	types.SERV.SetTaskDispatcher(taskDispatcher)

	journal := history.NewService(ft, dbService)
	if journal == nil {
		panic("Cannot initialize journal")
	}
	err = journal.Init(dbService)
	if err != nil {
		panic(err)
	}

	ft.SetJournal(journal)
	go journal.JournalWorker()
	go journal.FileWatcher()

	requester := routes.NewRequester()

	sw.Lap("Init services")

	localServer := instanceService.GetLocal()
	if localServer == nil {
		panic("Local server not initialized")
	}

	if localServer.ServerRole() == types.Backup {
		checkCoreExists(requester, sw)
	}

	caster := routes.NewCaster()
	caster.Disable()
	types.SERV.SetCaster(caster)

	// Enable the worker pool held by the task tracker
	// loading the filesystem might dispatch tasks,
	// so we have to start the pool first
	workerPool.Run()
	sw.Lap("Worker pool enabled")

	// Load filesystem
	util.Info.Println("Loading filesystem...")
	dataStore.FsInit(ft)
	sw.Lap("Initialized Filesystem")

	shareService := share.NewService()
	err = shareService.Init(dbService)
	if err != nil {
		panic(err)
	}
	types.SERV.SetShareService(shareService)

	err = dataStore.ClearTempDir(ft)
	if err != nil {
		panic(err)
	}
	sw.Lap("Clear tmp dir")

	err = dataStore.ClearTakeoutDir(ft)
	util.FailOnError(err, "Failed to clear takeout directory on startup")
	sw.Lap("Clear takeout dir")

	dataStore.InitApiKeyMap()
	sw.Lap("Api key map init")

	// The global broadcaster is disabled by default so all the
	// initial loading of the filesystem (that was just done above) doesn't
	// try to broadcast for every file that exists. So it must be enabled here
	// caster.DropBuffer()
	caster.Enable()
	sw.Lap("Global caster enabled")

	if localServer.ServerRole() == types.Backup {
		go dataProcess.BackupD(time.Minute, requester)
		sw.Lap("Init backup sleeper")
	}

	sw.Stop()
	sw.PrintResults(false)

	util.Info.Printf("Weblens loaded. %d files and %d medias\n", ft.Size(), mediaService.Size())

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
