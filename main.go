package weblens

import (
	"context"
	"os"
	"time"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore/database"
	"github.com/ethrousseau/weblens/api/fileTree"
	"github.com/ethrousseau/weblens/api/http"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/proxy"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/websocket"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		wlog.Warning.Println("Could not load .env file", err)
	}

	internal.LabelThread(
		func(_ context.Context) {
			go setup(internal.GetMongoDBName())
		}, "", "Setup",
	)

	internal.LabelThread(
		func(_ context.Context) {
			http.DoRoutes()
		}, "", "Router",
	)
	wlog.Info.Println("Weblens exited...")
}

func setupRecovery() {
	err := recover()
	if err != nil {
		wlog.ErrTrace(err.(error))
		wlog.ErrorCatcher.Println("WEBLENS STARTUP FAILED.")
		os.Exit(1)
	}
}

func setup(mongoName string) {
	defer setupRecovery()

	sw := internal.NewStopwatch("Initialization")

	if internal.IsDevMode() {
		wlog.DoDebug()
		wlog.Debug.Println("Starting weblens in development mode")
	} else {
	}
	gin.SetMode(gin.ReleaseMode)

	sw.Lap()

	/* Database connection */
	db := database.New(internal.GetMongoURI(), mongoName)
	sw.Lap("Connect to Mongo")

	/* Access Service */
	accessService := NewAccessService(db.Collection("apiKeys"))
	err := accessService.Init()
	if err != nil {
		panic(err)
	}
	sw.Lap("Init access service")

	/* Instance Service */
	instanceService := NewInstanceService(accessService, db.Collection("servers"))
	err = instanceService.Init()
	if err != nil {
		panic(err)
	}
	types.SERV.SetInstance(instanceService)
	localServer := instanceService.GetLocal()
	if localServer == nil {
		panic("Local server not initialized")
	}
	sw.Lap("Init instance service")

	/* Client Manager */
	clientService := websocket.NewClientManager()
	types.SERV.SetClientService(clientService)
	sw.Lap("Init client service")

	/* If server is backup server, connect to core server */
	if localServer.ServerRole() == BackupServer {
		core := instanceService.GetCore()
		if core == nil {
			panic("could not get core")
		}

		err = websocket.WebsocketToCore(core)
		if err != nil {
			panic(err)
		}

		var coreAddr string
		coreAddr, err = core.GetAddress()
		if err != nil {
			panic(err)
		}

		if coreAddr == "" || instanceService.GetCore().GetUsingKey() == "" {
			panic("could not get core address or key")
		}

		proxyStore := proxy.NewProxyStore(coreAddr, instanceService.GetCore().GetUsingKey())
		sw.Lap("Init proxy store")

		proxyStore.Init(localStore)
		types.SERV.SetStore(proxyStore)
		localStore = proxyStore
	}

	/* User Service */
	userService := NewUserService()
	err = userService.Init(localStore)
	if err != nil {
		panic(err)
	}
	// types.SERV.SetUserService(userService)
	http.UserService = userService
	sw.Lap("Init user service")


	/* Once here, we can actually let the router start, this is the minimally
	acceptable state to allow incoming HTTP, i.e. types.SERV.MinimallyReady() == true */

	/* Share Service */
	shareService := NewAlbumService()
	err = shareService.Init(localStore)
	if err != nil {
		panic(err)
	}
	types.SERV.SetShareService(shareService)
	sw.Lap("Init share service")

	/* FileTree Service */
	fTree := fileTree.NewFileTree(internal.GetMediaRootPath(), "MEDIA", userService)
	types.SERV.SetFileTree(fTree)
	err = fTree.Init(localStore)
	if err != nil {
		panic(err)
	}
	sw.Lap("Init file tree service")

	/* Media type Service */
	mediaTypeServ := NewTypeService()
	/* Media Service */
	mediaService := NewMediaService(mediaTypeServ)
	err = mediaService.Init(localStore)
	if err != nil {
		panic(err)
	}
	types.SERV.SetMediaRepo(mediaService)
	sw.Lap("Init media service")

	/* Album Service */
	albumService := NewUserService()
	err = albumService.Init(localStore)
	if err != nil {
		panic(err)
	}
	types.SERV.SetAlbumService(albumService)
	sw.Lap("Init album service")

	/* Journal Service */
	journal := fileTree.NewJournalService(fTree)
	if journal == nil {
		panic("Cannot initialize journal")
	}
	err = journal.Init(localStore)
	if err != nil {
		panic(err)
	}
	sw.Lap("Init journal service")

	/* Give journal to file tree, and start workers */
	fTree.SetJournal(journal)
	internal.LabelThread(
		func(_ context.Context) {
			go journal.JournalWorker()
		}, "", "Journal Worker",
	)

	internal.LabelThread(
		func(_ context.Context) {
			go journal.FileWatcher()
		}, "", "File Watcher",
	)

	/* The global broadcaster is created and disabled here so that we don't
	read information about files before they are ready to be accessed */
	caster := websocket.NewCaster()
	caster.Disable()
	types.SERV.SetCaster(caster)

	/* Enable the worker pool held by the task tracker
	loading the filesystem might dispatch tasks,
	so we have to start the pool first */
	workerPool, taskDispatcher := dataProcess.NewWorkerPool(internal.GetWorkerCount(), fTree, userService)
	types.SERV.SetWorkerPool(workerPool)
	types.SERV.SetTaskDispatcher(taskDispatcher)
	workerPool.Run()
	sw.Lap("Worker pool enabled")

	if InstanceService.GetLocal().ServerRole() != types.Initialization {
		/* Load filesystem into tree */
		wlog.Debug.Println("Loading filesystem...")
		hashCaster := websocket.NewCaster()
		err = fTree.InitMediaRoot(hashCaster)
		if err != nil {
			panic(err)
		}
		sw.Lap("Loaded Filesystem")
	}

	/* If we are on a backup server, launch the backup daemon */
	if localServer.ServerRole() == BackupServer {
		go dataProcess.BackupD(time.Hour)
	}

	caster.Enable()

	sw.Stop()
	sw.PrintResults(false)
	wlog.Info.Printf(
		"Weblens loaded in %s. %d files and %d medias\n", sw.GetTotalTime(false), fTree.Size(),
		mediaService.Size(),
	)

	InstanceService.RemoveLoading("all")

}
