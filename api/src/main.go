package main

import (
	"context"
	"errors"
	"fmt"
	"os"
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
	"github.com/ethrousseau/weblens/api/util/wlog"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		wlog.Warning.Println("Could not load .env file", err)
	}

	util.LabelThread(
		func(_ context.Context) {
			go setup(util.GetMongoDBName())
		}, "", "Setup",
	)

	util.LabelThread(
		func(_ context.Context) {
			routes.DoRoutes()
		}, "", "Router",
	)
	wlog.Info.Println("Weblens exited...")
}

func setupRecovery() {
	err := recover()
	if err != nil {
		switch err := err.(type) {
		case types.WeblensError:
			wlog.ErrTrace(err)
		case error:
			wlog.ErrTrace(err)
		default:
			wlog.ErrTrace(errors.New(fmt.Sprintln("Recovered unexpected", err)))
		}
		wlog.ErrorCatcher.Println("WEBLENS STARTUP FAILED.")
		os.Exit(1)
	}
}

func setup(mongoName string) {
	defer setupRecovery()

	sw := util.NewStopwatch("Initialization")

	if util.IsDevMode() {
		wlog.DoDebug()
		wlog.Debug.Println("Starting weblens in development mode")
	} else {
	}
	gin.SetMode(gin.ReleaseMode)

	sw.Lap()

	/* Get database connection */
	localStore := database.New(util.GetMongoURI(), mongoName)
	types.SERV.SetStore(localStore)
	sw.Lap("Create local store controller")

	/* Instance Service */
	instanceService := instance.NewService()
	err := instanceService.Init(localStore)
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
	clientService := routes.NewClientManager()
	types.SERV.SetClientService(clientService)
	sw.Lap("Init client service")

	/* If server is backup server, connect to core server */
	if localServer.ServerRole() == types.Backup {
		core := types.SERV.InstanceService.GetCore()
		if core == nil {
			panic("could not get core")
		}

		err = routes.WebsocketToCore(core)
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
	userService := user.NewService()
	err = userService.Init(localStore)
	if err != nil {
		panic(err)
	}
	types.SERV.SetUserService(userService)
	sw.Lap("Init user service")

	/* Once here, we can actually let the router start, this is the minimally
	acceptable state to allow incoming HTTP, i.e. types.SERV.MinimallyReady() == true */

	/* Share Service */
	shareService := share.NewService()
	err = shareService.Init(localStore)
	if err != nil {
		panic(err)
	}
	types.SERV.SetShareService(shareService)
	sw.Lap("Init share service")

	/* FileTree Service */
	ft := filetree.NewFileTree(util.GetMediaRootPath(), "MEDIA")
	types.SERV.SetFileTree(ft)
	err = ft.Init(localStore)
	if err != nil {
		panic(err)
	}
	sw.Lap("Init file tree service")

	/* Media type Service */
	mediaTypeServ := media.NewTypeService()
	/* Media Service */
	mediaService := media.NewRepo(mediaTypeServ)
	err = mediaService.Init(localStore)
	if err != nil {
		panic(err)
	}
	types.SERV.SetMediaRepo(mediaService)
	sw.Lap("Init media service")

	/* Album Service */
	albumService := album.NewService()
	err = albumService.Init(localStore)
	if err != nil {
		panic(err)
	}
	types.SERV.SetAlbumService(albumService)
	sw.Lap("Init album service")

	/* Access Service */
	accessService := dataStore.NewAccessService()
	err = accessService.Init(localStore)
	if err != nil {
		panic(err)
	}
	types.SERV.SetAccessService(accessService)
	sw.Lap("Init access service")

	/* Journal Service */
	journal := history.NewService(ft)
	if journal == nil {
		panic("Cannot initialize journal")
	}
	err = journal.Init(localStore)
	if err != nil {
		panic(err)
	}
	sw.Lap("Init journal service")

	/* Give journal to file tree, and start workers */
	ft.SetJournal(journal)
	util.LabelThread(
		func(_ context.Context) {
			go journal.JournalWorker()
		}, "", "Journal Worker",
	)

	util.LabelThread(
		func(_ context.Context) {
			go journal.FileWatcher()
		}, "", "File Watcher",
	)

	/* The global broadcaster is created and disabled here so that we don't
	read information about files before they are ready to be accessed */
	caster := routes.NewCaster()
	caster.Disable()
	types.SERV.SetCaster(caster)

	/* Enable the worker pool held by the task tracker
	loading the filesystem might dispatch tasks,
	so we have to start the pool first */
	workerPool, taskDispatcher := dataProcess.NewWorkerPool(util.GetWorkerCount())
	types.SERV.SetWorkerPool(workerPool)
	types.SERV.SetTaskDispatcher(taskDispatcher)
	workerPool.Run()
	sw.Lap("Worker pool enabled")

	if types.SERV.InstanceService.GetLocal().ServerRole() != types.Initialization {
		/* Load filesystem into tree */
		wlog.Debug.Println("Loading filesystem...")
		hashCaster := routes.NewCaster()
		err = dataStore.InitMediaRoot(ft, hashCaster)
		if err != nil {
			panic(err)
		}
		sw.Lap("Loaded Filesystem")
	}

	/* If we are on a backup server, launch the backup daemon */
	if localServer.ServerRole() == types.Backup {
		go dataProcess.BackupD(time.Hour)
	}

	caster.Enable()

	sw.Stop()
	sw.PrintResults(false)
	wlog.Info.Printf(
		"Weblens loaded in %s. %d files and %d medias\n", sw.GetTotalTime(false), ft.Size(),
		mediaService.Size(),
	)

	types.SERV.InstanceService.RemoveLoading("all")

}
