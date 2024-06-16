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
	"github.com/ethrousseau/weblens/api/dataStore/media"
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
		panic(err)
	}

	if util.IsDevMode() {
		util.Debug.Println("Initializing weblens in development mode")
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	sw.Lap()

	// Gather global services
	dbService := database.New()
	fileTree := filetree.NewFileTree()
	caster := routes.NewBufferedCaster()
	mediaTypeServ := media.NewTypeService()
	mediaServ := media.NewRepo(mediaTypeServ)
	albumServ := album.NewService(dbService)
	// dataProcess.SetControllers(mediaServ)
	sw.Lap("Init controllers map")

	workerPool, taskDispatcher := dataProcess.NewWorkerPool(runtime.NumCPU() - 2)

	journal := history.NewJournalService(fileTree, dbService)
	if journal == nil {
		panic("Cannot initialize journal")
	}

	fileTree.SetJournal(journal)

	go journal.JournalWorker()
	go journal.FileWatcher()

	requester := routes.NewRequester()
	srvInfo := dataStore.GetServerInfo()
	if srvInfo.ServerRole() == types.Backup {
		checkCoreExists(requester, sw)
	}

	clientManager := routes.NewClientManager()
	routes.SetControllers(fileTree, mediaServ, caster, clientManager, taskDispatcher, requester, albumServ)
	history.SetHistoryControllers(fileTree, dbService)

	err = dataStore.ClearTempDir(fileTree)
	util.FailOnError(err, "Failed to clear temporary directory on startup")
	sw.Lap("Clear tmp dir")

	err = dataStore.ClearTakeoutDir(fileTree)
	util.FailOnError(err, "Failed to clear takeout directory on startup")
	sw.Lap("Clear takeout dir")

	store := dataStore.NewStore(requester)

	util.Info.Println("Loading users...")
	err = store.LoadUsers(fileTree)
	util.FailOnError(err, "Failed to load users")
	sw.Lap("Users init")

	// Enable the worker pool held by the task tracker
	// loading the filesystem might dispatch tasks,
	// so we have to start the pool first
	workerPool.Run()
	sw.Lap("Worker pool enabled")

	// Load filesystem
	util.Info.Println("Loading filesystem...")
	dataStore.FsInit(fileTree, dbService, taskDispatcher, caster)
	sw.Lap("FS init")
	util.Info.Println("Initialized Filesystem")

	err = media.MediaInit()
	if err != nil {
		panic(err)
	}
	sw.Lap("Media init")
	util.Info.Println("Initialized Media Map")

	dataStore.VerifyAlbumsMedia()
	sw.Lap("Albums init")

	dataStore.LoadAllShares(fileTree)
	sw.Lap("Shares init")

	dataStore.InitApiKeyMap()
	sw.Lap("Api key map init")

	// The global broadcaster is disabled by default so all the
	// initial loading of the filesystem (that was just done above) doesn't
	// try to broadcast for every file that exists. So it must be enabled here
	caster.DropBuffer()
	sw.Lap("Global caster enabled")

	if srvInfo.ServerRole() == types.Backup {
		go dataProcess.BackupD(time.Minute, requester)
		sw.Lap("Init backup sleeper")
	}

	sw.Stop()
	sw.PrintResults(false)

	util.Info.Printf("Weblens loaded. %d files and %d medias\n", fileTree.Size(), mediaServ.Size())

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
