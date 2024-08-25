package main

import (
	"context"
	"os"
	"time"

	"github.com/ethrousseau/weblens/backup"
	"github.com/ethrousseau/weblens/comm"
	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/jobs"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/models/service"
	"github.com/ethrousseau/weblens/task"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	} else {
	}
	gin.SetMode(gin.ReleaseMode)

	sw.Lap()

	/* Database connection */
	db := connectToMongo(internal.GetMongoURI(), mongoName)
	sw.Lap("Connect to Mongo")

	/* Access Service */
	accessService := service.NewAccessService(db.Collection("apiKeys"))
	err := accessService.Init()
	if err != nil {
		panic(err)
	}
	comm.AccessService = accessService
	sw.Lap("Init access service")

	/* Instance Service */
	instanceService := service.NewInstanceService(accessService, db.Collection("servers"))
	err = instanceService.Init()
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
	workerPool.RegisterJob(models.WriteFileTask, jobs.HandleFileUploads)
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

	fileService, err := service.NewFileService(
		mediaFileTree, cachesTree, userService, accessService,
		db.Collection("trash"),
	)
	if err != nil {
		panic(err)
	}
	comm.FileService = fileService
	sw.Lap("Init file tree service")

	/* Client Manager */
	clientService := comm.NewClientManager(fileService, workerPool)
	comm.ClientService = clientService
	sw.Lap("Init client service")

	/* Media type Service */
	mediaTypeServ := models.NewTypeService()
	/* Media Service */
	mediaService := service.NewMediaService(fileService, nil, mediaTypeServ, db.Collection("media"))
	err = mediaService.Init()
	if err != nil {
		panic(err)
	}
	comm.MediaService = mediaService
	sw.Lap("Init media service")

	/* Album Service */
	albumService := service.NewAlbumService(db.Collection("albums"), mediaService)
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

const maxRetries = 5

func connectToMongo(mongoUri, mongoDbName string) *mongo.Database {
	clientOptions := options.Client().ApplyURI(mongoUri).SetTimeout(time.Second)
	var err error
	mongoc, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}

	retries := 0
	for retries < maxRetries {
		err = mongoc.Ping(context.Background(), nil)
		if err == nil {
			break
		}
		log.Warning.Printf("Failed to connect to mongo, trying %d more time(s)", maxRetries-retries)
		time.Sleep(time.Second * 1)
		retries++
	}
	if err != nil {
		log.Error.Printf("Failed to connect to database after %d retries", maxRetries)
		panic(err)
	}

	log.Debug.Println("Connected to mongo")

	return mongoc.Database(mongoDbName)
}

func mainRecovery(msg string) {
	err := recover()
	if err != nil {
		log.ErrTrace(err.(error))
		log.ErrorCatcher.Println(msg)
		os.Exit(1)
	}
}
