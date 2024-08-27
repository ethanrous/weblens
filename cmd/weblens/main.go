package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ethrousseau/weblens/backup"
	"github.com/ethrousseau/weblens/comm"
	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/jobs"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/models/service"
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

	internal.LabelThread(
		func(_ context.Context) {
			startup(internal.GetMongoDBName())
		}, "", "Setup",
	)

	for {
		internal.LabelThread(
			func(_ context.Context) {
				comm.DoRoutes()
			}, "", "Router",
		)
	}
}

func _main() {

}

func startup(mongoName string) {
	defer mainRecovery("WEBLENS STARTUP FAILED")

	sw := internal.NewStopwatch("Initialization")

	if internal.IsDevMode() {
		log.DoDebug()
		log.Debug.Println("Starting weblens in development mode")

		metricsServer := http.Server{
			Addr:     "localhost:2112",
			ErrorLog: log.ErrorCatcher,
			Handler:  promhttp.Handler(),
		}
		go metricsServer.ListenAndServe()
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

	// TODO
	/* If server is backup server, connect to core server */
	// if localServer.ServerRole() == BackupServer {
	// 	core := instanceService.GetCore()
	// 	if core == nil {
	// 		panic("could not get core")
	// 	}
	//
	// 	err = websocket.WebsocketToCore(core)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	//
	// 	var coreAddr string
	// 	coreAddr, err = core.GetAddress()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	//
	// 	if coreAddr == "" || instanceService.GetCore().GetUsingKey() == "" {
	// 		panic("could not get core address or key")
	// 	}
	//
	// 	proxyStore := proxy.NewProxyStore(coreAddr, instanceService.GetCore().GetUsingKey())
	// 	sw.Lap("Init proxy store")
	//
	// 	proxyStore.Init(localStore)
	// 	types.SERV.SetStore(proxyStore)
	// 	localStore = proxyStore
	// }

	/* User Service */
	userService := service.NewUserService(db.Collection("users"))
	err = userService.Init()
	if err != nil {
		panic(err)
	}
	comm.UserService = userService
	sw.Lap("Init user service")

	/* Once here, we can actually let the router start, this is the minimally
	acceptable state to allow incoming HTTP, i.e. types.SERV.MinimallyReady() == true */

	internal.LabelThread(
		func(_ context.Context) {
			go comm.DoRoutes()
		}, "", "Router",
	)

	/* Enable the worker pool held by the task tracker
	loading the filesystem might dispatch tasks,
	so we have to start the pool first */
	workerPool := task.NewWorkerPool(internal.GetWorkerCount())

	workerPool.RegisterJob(models.ScanDirectoryTask, jobs.ScanDirectory)
	workerPool.RegisterJob(models.ScanFileTask, jobs.ScanFile)
	workerPool.RegisterJob(models.MoveFileTask, jobs.MoveFile)
	workerPool.RegisterJob(models.UploadFilesTask, jobs.HandleFileUploads)
	workerPool.RegisterJob(models.CreateZipTask, jobs.CreateZip)
	workerPool.RegisterJob(models.GatherFsStatsTask, jobs.GatherFilesystemStats)
	workerPool.RegisterJob(models.BackupTask, backup.DoBackup)
	workerPool.RegisterJob(models.HashFile, jobs.HashFile)
	workerPool.RegisterJob(models.CopyFileFromCore, backup.CopyFileFromCore)

	comm.TaskService = workerPool
	workerPool.Run()
	sw.Lap("Worker pool enabled")

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

	// Hasher
	hasher := models.NewHasher(workerPool)

	/* FileTree Service */
	mediaFileTree, err := fileTree.NewFileTree(internal.GetMediaRootPath(), "MEDIA", hasher, mediaJournal)
	if err != nil {
		panic(err)
	}

	hollowJournal := fileTree.NewHollowJournalService()
	hollowHasher := models.NewHollowHasher()
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
	typeJson, err := os.Open(filepath.Join(internal.GetConfigDir(), "mediaType.json"))
	if err != nil {
		panic(err)
	}
	defer func(typeJson *os.File) {
		err := typeJson.Close()
		if err != nil {
			panic(err)
		}
	}(typeJson)

	typesBytes, err := io.ReadAll(typeJson)
	marshMap := map[string]models.MediaType{}
	err = json.Unmarshal(typesBytes, &marshMap)
	if err != nil {
		panic(err)
	}
	mediaTypeServ := models.NewTypeService(marshMap)
	/* Media Service */
	mediaService, err := service.NewMediaService(fileService, mediaTypeServ, db.Collection("media"))
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

	/* Client Manager */
	clientService := comm.NewClientManager(fileService, workerPool)
	comm.ClientService = clientService
	sw.Lap("Init client service")

	/* Album Service */
	albumService := service.NewAlbumService(db.Collection("albums"), mediaService, shareService)
	err = albumService.Init()
	if err != nil {
		panic(err)
	}
	comm.AlbumService = albumService
	sw.Lap("Init album service")

	/* Give journal to file tree, and start workers */

	/* If we are on a backup server, launch the backup daemon */
	if localServer.ServerRole() == models.BackupServer {
		go backup.BackupD(time.Hour, instanceService, workerPool)
	}

	/* Base / global caster */
	caster := comm.NewSimpleCaster(clientService)
	comm.Caster = caster

	sw.Stop()
	sw.PrintResults(false)
	log.Info.Printf(
		"Weblens loaded in %s. %d files and %d medias\n", sw.GetTotalTime(false), fileService.Size(),
		mediaService.Size(),
	)

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
