package main

import (
	"net/http"
	"os"
	"time"

	"github.com/ethrousseau/weblens/comm"
	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/jobs"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/service"
	"github.com/ethrousseau/weblens/service/mock"
	"github.com/ethrousseau/weblens/task"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	defer mainRecovery("WEBLENS ENCOUNTERED AN UNRECOVERABLE ERROR")

	err := godotenv.Load("./config/.env")
	if err != nil {
		log.Warning.Println("Could not load .env file", err)
	}

	startup(internal.GetMongoDBName())
	comm.DoRoutes()
}

func startup(mongoName string) {
	defer mainRecovery("WEBLENS STARTUP FAILED")

	log.Info.Println("Starting Weblens")

	sw := internal.NewStopwatch("Initialization")

	if internal.IsDevMode() {
		log.DoDebug()
		log.Debug.Println("Starting Weblens in debug mode")

		metricsServer := http.Server{
			Addr:     "localhost:2112",
			ErrorLog: log.ErrorCatcher,
			Handler:  promhttp.Handler(),
		}
		go log.ErrTrace(metricsServer.ListenAndServe())
	}

	gin.SetMode(gin.ReleaseMode)

	sw.Lap()

	/* Database connection */
	db := database.ConnectToMongo(internal.GetMongoURI(), mongoName)
	sw.Lap("Connect to Mongo")

	/* Instance Service */
	instanceService := service.NewInstanceService(db.Collection("servers"))
	err := instanceService.Init()
	if err != nil {
		panic(err)
	}
	comm.InstanceService = instanceService
	localServer := instanceService.GetLocal()
	if localServer == nil {
		panic("Local server not initialized")
	}
	sw.Lap("Init instance service")

	/* User Service */
	userService := service.NewUserService(db.Collection("users"))
	err = userService.Init()
	if err != nil {
		panic(err)
	}
	comm.UserService = userService
	sw.Lap("Init user service")

	/* Enable the worker pool held by the task tracker
	loading the filesystem might dispatch tasks,
	so we have to start the pool first */
	workerPool := task.NewWorkerPool(internal.GetWorkerCount())

	workerPool.RegisterJob(models.ScanDirectoryTask, jobs.ScanDirectory)
	workerPool.RegisterJob(models.ScanFileTask, jobs.ScanFile)
	// workerPool.RegisterJob(models.MoveFileTask, jobs.MoveFiles)
	workerPool.RegisterJob(models.UploadFilesTask, jobs.HandleFileUploads)
	workerPool.RegisterJob(models.CreateZipTask, jobs.CreateZip)
	workerPool.RegisterJob(models.GatherFsStatsTask, jobs.GatherFilesystemStats)
	workerPool.RegisterJob(models.BackupTask, jobs.DoBackup)
	workerPool.RegisterJob(models.HashFileTask, jobs.HashFile)
	workerPool.RegisterJob(models.CopyFileFromCoreTask, jobs.CopyFileFromCore)

	comm.TaskService = workerPool
	workerPool.Run()
	sw.Lap("Worker pool enabled")

	/* Client Manager */
	clientService := service.NewClientManager(nil, workerPool, instanceService)
	comm.ClientService = clientService
	sw.Lap("Init client service")

	/* If server is backup server, connect to core server and launch backup daemon */
	if localServer.ServerRole() == models.BackupServer {
		core := instanceService.GetCore()
		if core == nil {
			panic("could not get core")
		}

		err = comm.WebsocketToCore(core, clientService)
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
	}

	/* Once here, we can actually let the router start, this is the minimally
	acceptable state to allow incoming HTTP, i.e. types.SERV.MinimallyReady() == true */

	go comm.DoRoutes()

	/* Share Service */
	shareService := service.NewShareService(db.Collection("shares"))
	err = shareService.Init()
	if err != nil {
		panic(err)
	}
	comm.ShareService = shareService
	sw.Lap("Init share service")

	/* Journal Service */
	mediaJournal, err := fileTree.NewJournalService(
		db.Collection("fileHistory"),
		string(instanceService.GetLocal().ServerId()),
	)
	if err != nil {
		panic(err)
	}
	sw.Lap("Init journal service")

	/* Baseic global caster */
	caster := models.NewSimpleCaster(clientService)
	comm.Caster = caster

	// Hasher
	hasher := models.NewHasher(workerPool, caster)

	/* FileTree Service */
	mediaFileTree, err := fileTree.NewFileTree(internal.GetMediaRootPath(), "MEDIA", hasher, mediaJournal)
	if err != nil {
		panic(err)
	}

	hollowJournal := mock.NewHollowJournalService()
	hollowHasher := mock.NewMockHasher()
	cachesTree, err := fileTree.NewFileTree(internal.GetCacheRoot(), "CACHES", hollowHasher, hollowJournal)
	if err != nil {
		panic(err)
	}
	sw.Lap("Init file trees")

	fileService, err := service.NewFileService(
		mediaFileTree, cachesTree, userService, nil, nil,
		db.Collection("trash"),
	)
	if err != nil {
		panic(err)
	}
	comm.FileService = fileService
	sw.Lap("Init file service")

	/* Media type Service */
	// Only from config file, for now
	marshMap := map[string]models.MediaType{}
	internal.ReadTypesConfig(&marshMap)
	mediaTypeServ := models.NewTypeService(marshMap)
	/* Media Service */
	mediaService, err := service.NewMediaService(
		fileService, mediaTypeServ, &mock.MockAlbumService{},
		db.Collection("media"),
	)
	if err != nil {
		panic(err)
	}
	comm.MediaService = mediaService
	sw.Lap("Init media service")

	/* Access Service */
	accessService := service.NewAccessService(fileService, db.Collection("apiKeys"))
	err = accessService.Init()
	if err != nil {
		panic(err)
	}
	comm.AccessService = accessService
	sw.Lap("Init access service")

	fileService.SetAccessService(accessService)
	fileService.SetMediaService(mediaService)
	clientService.SetFileService(fileService)

	/* Album Service */
	albumService := service.NewAlbumService(db.Collection("albums"), mediaService, shareService)
	err = albumService.Init()
	if err != nil {
		panic(err)
	}
	comm.AlbumService = albumService
	sw.Lap("Init album service")

	mediaService.AlbumService = albumService

	sw.Stop()
	sw.PrintResults(false)
	log.Info.Printf(
		"Weblens loaded in %s. %d files and %d medias\n", sw.GetTotalTime(false), fileService.Size(),
		mediaService.Size(),
	)

	if localServer.ServerRole() == models.BackupServer {
		go jobs.BackupD(time.Hour, instanceService, workerPool, fileService, userService, clientService, caster)
	}

	instanceService.RemoveLoading("all")
}

func mainRecovery(msg string) {
	err := recover()
	if err != nil {
		log.ErrTrace(err.(error))
		log.ErrorCatcher.Println(msg)
		os.Exit(1)
	}
}
