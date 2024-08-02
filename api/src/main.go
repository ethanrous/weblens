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
	"github.com/ethrousseau/weblens/api/routes/proxy"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const retries = 10

func main() {
	go setup(util.GetMongoDBName())
	routes.DoRoutes()
	util.Info.Println("Weblens exited...")
}

func setup(mongoName string) {
	sw := util.NewStopwatch("Initialization")
	err := godotenv.Load()
	if err != nil {
		util.ShowErr(err)
	}

	if util.IsDevMode() {
		util.Debug.Println("Starting weblens in development mode")
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	sw.Lap()

	// Gather global services

	localStore := database.New(util.GetMongoURI(), mongoName)
	types.SERV.SetStore(localStore)

	sw.Lap("Create local store controller")

	instanceService := instance.NewService()
	err = instanceService.Init(localStore)
	if err != nil {
		panic(err)
	}
	types.SERV.SetInstance(instanceService)

	sw.Lap("Init instance service")

	localServer := instanceService.GetLocal()
	if localServer == nil {
		panic("Local server not initialized")
	}

	userService := user.NewService()

	var proxyStore *proxy.ProxyStore
	if localServer.ServerRole() == types.Backup {
		coreAddr, err := localServer.GetCoreAddress()
		if err != nil {
			panic(err)
		}
		proxyStore = proxy.NewProxyStore(coreAddr, localServer.GetUsingKey())
		sw.Lap("Init proxy store")

		proxyStore.Init(localStore)
		types.SERV.SetStore(proxyStore)

		err = userService.Init(proxyStore)
	} else {
		err = userService.Init(localStore)
	}
	if err != nil {
		panic(err)
	}
	types.SERV.SetUserService(userService)
	sw.Lap("Init user service")

	clientService := routes.NewClientManager()

	types.SERV.SetClientService(clientService)
	sw.Lap("Init client service")

	/* Once here, we can actually let the router start, this is the minimally
	acceptable state to allow incoming HTTP, i.e. types.SERV.MinimallyReady() == true */

	shareService := share.NewService()
	err = shareService.Init(localStore)
	if err != nil {
		panic(err)
	}
	types.SERV.SetShareService(shareService)
	sw.Lap("Init share service")

	ft := filetree.NewFileTree(util.GetMediaRootPath(), "MEDIA")
	types.SERV.SetFileTree(ft)
	if localServer.ServerRole() == types.Backup {
		err = ft.Init(proxyStore)
	} else {
		err = ft.Init(localStore)
	}
	if err != nil {
		panic(err)
	}
	sw.Lap("Init file tree service")

	mediaTypeServ := media.NewTypeService()
	mediaService := media.NewRepo(mediaTypeServ)
	if localServer.ServerRole() == types.Backup {
		err = mediaService.Init(proxyStore)
	} else {
		err = mediaService.Init(localStore)
	}
	if err != nil {
		panic(err)
	}
	types.SERV.SetMediaRepo(mediaService)
	sw.Lap("Init media service")

	albumService := album.NewService()
	err = albumService.Init(localStore)
	if err != nil {
		panic(err)
	}
	types.SERV.SetAlbumService(albumService)

	sw.Lap("Init album service")

	accessService := dataStore.NewAccessService()
	err = accessService.Init(localStore)
	if err != nil {
		panic(err)
	}
	types.SERV.SetAccessService(accessService)
	sw.Lap("Init access service")

	journal := history.NewService(ft)
	if journal == nil {
		panic("Cannot initialize journal")
	}
	if localServer.ServerRole() == types.Backup {
		err = journal.Init(proxyStore)
	} else {
		err = journal.Init(localStore)
	}
	if err != nil {
		panic(err)
	}

	ft.SetJournal(journal)
	go journal.JournalWorker()
	go journal.FileWatcher()
	sw.Lap("Init journal service")

	requester := routes.NewRequester()
	types.SERV.SetRequester(requester)

	if localServer.ServerRole() == types.Backup {
		checkCoreExists(requester, sw)
	}

	// The global broadcaster is created and disabled here so that we don't
	// read information about files before they are ready to be accessed
	caster := routes.NewCaster()
	caster.Disable()
	types.SERV.SetCaster(caster)

	// Enable the worker pool held by the task tracker
	// loading the filesystem might dispatch tasks,
	// so we have to start the pool first
	workerPool, taskDispatcher := dataProcess.NewWorkerPool(runtime.NumCPU() - 2)
	types.SERV.SetWorkerPool(workerPool)
	types.SERV.SetTaskDispatcher(taskDispatcher)
	workerPool.Run()
	sw.Lap("Worker pool enabled")

	// Load filesystem
	util.Info.Println("Loading filesystem...")
	hashCaster := routes.NewCaster()
	err = dataStore.InitMediaRoot(ft, hashCaster)
	if err != nil {
		panic(err)
	}
	sw.Lap("Loaded Filesystem")

	err = dataStore.ClearTempDir(ft)
	if err != nil {
		panic(err)
	}
	sw.Lap("Clear tmp dir")

	err = dataStore.ClearTakeoutDir(ft)
	if err != nil {
		panic(err)
	}
	sw.Lap("Clear takeout dir")

	// If we are on a backup server, launch the backup daemon
	if localServer.ServerRole() == types.Backup {
		go dataProcess.BackupD(time.Hour)
	}

	caster.Enable()

	sw.Stop()
	sw.PrintResults(false)

	types.SERV.InstanceService.RemoveLoading("all")

	util.Info.Printf("Weblens loaded. %d files and %d medias\n", ft.Size(), mediaService.Size())
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
