package main

import (
	"net/http"
	"os"
	"time"

	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/fileTree"
	. "github.com/ethrousseau/weblens/http"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/env"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/jobs"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/service"
	"github.com/ethrousseau/weblens/service/mock"
	"github.com/ethrousseau/weblens/task"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var server *Server
var services = &models.ServicePack{}

func main() {
	config, err := env.ReadConfig("", os.Getenv("CONFIG_NAME"))
	if err != nil {
		panic(err)
	}

	defer mainRecovery("WEBLENS ENCOUNTERED AN UNRECOVERABLE ERROR")
	log.Info.Println("Starting Weblens")

	if config["logLevel"].(string) == "debug" {
		log.SetLogLevel(log.DEBUG)
		log.Debug.Println("Starting Weblens in debug mode")

		metricsServer := http.Server{
			Addr:     "localhost:2112",
			ErrorLog: log.ErrorCatcher,
			Handler:  promhttp.Handler(),
		}
		go func() { log.ErrTrace(metricsServer.ListenAndServe()) }()
	} else {
	}
	gin.SetMode(gin.ReleaseMode)

	server = NewServer(config["routerHost"].(string), config["routerPort"].(string), services)
	go startup(config, services, server)

	server.Start()
}

func startup(config map[string]any, pack *models.ServicePack, srv *Server) {
	defer mainRecovery("WEBLENS STARTUP FAILED")

	log.Trace.Println("Beginning service setup")

	sw := internal.NewStopwatch("Initialization")
	sw.Lap()

	/* Database connection */
	db, err := database.ConnectToMongo(config["mongodbUri"].(string), config["mongodbName"].(string))
	if err != nil {
		panic(err)
	}
	sw.Lap("Connect to Mongo")

	/* Instance Service */
	instanceService, err := service.NewInstanceService(db.Collection("servers"))
	if err != nil {
		panic(err)
	}
	pack.InstanceService = instanceService
	localServer := instanceService.GetLocal()
	sw.Lap("Init instance service")

	/* User Service */
	userService, err := service.NewUserService(db.Collection("users"))
	if err != nil {
		panic(err)
	}
	pack.UserService = userService
	sw.Lap("Init user service")

	wpLogLevel := 0
	if env.IsDevMode() {
		wpLogLevel = 1
	}
	/* Enable the worker pool held by the task tracker
	loading the filesystem might dispatch tasks,
	so we have to start the pool first */
	workerPool := task.NewWorkerPool(env.GetWorkerCount(), wpLogLevel)

	workerPool.RegisterJob(models.ScanDirectoryTask, jobs.ScanDirectory)
	workerPool.RegisterJob(models.ScanFileTask, jobs.ScanFile)
	workerPool.RegisterJob(models.UploadFilesTask, jobs.HandleFileUploads)
	workerPool.RegisterJob(models.CreateZipTask, jobs.CreateZip)
	workerPool.RegisterJob(models.GatherFsStatsTask, jobs.GatherFilesystemStats)
	workerPool.RegisterJob(models.BackupTask, jobs.DoBackup)
	workerPool.RegisterJob(models.HashFileTask, jobs.HashFile)
	workerPool.RegisterJob(models.CopyFileFromCoreTask, jobs.CopyFileFromCore)

	pack.TaskService = workerPool
	workerPool.Run()
	sw.Lap("Worker pool enabled")

	/* Client Manager */
	clientService := service.NewClientManager(nil, workerPool, instanceService)
	pack.ClientService = clientService
	sw.Lap("Init client service")

	log.Debug.Printf("Local server is %s", localServer.ServerRole())

	if localServer.ServerRole() == models.InitServer {
		srv.UseInit()
	} else {
		srv.UseApi()
	}

	/* If server is backup server, connect to core server and launch backup daemon */
	if localServer.ServerRole() == models.BackupServer {
		core := instanceService.GetCore()
		if core == nil {
			panic(werror.Errorf("Could not find core instance"))
		}

		pack.ClientService = clientService
		err = WebsocketToCore(core, clientService)
		if err != nil {
			panic(err)
		}

		var coreAddr string
		coreAddr, err = core.GetAddress()
		if err != nil {
			panic(err)
		}

		if coreAddr == "" || instanceService.GetCore().GetUsingKey() == "" {
			panic(werror.Errorf("could not get core address or key"))
		}

	} else if localServer.ServerRole() == models.CoreServer {
		srv.UseCore()
	}

	/* Share Service */
	shareService, err := service.NewShareService(db.Collection("shares"))
	if err != nil {
		panic(err)
	}
	pack.ShareService = shareService
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

	/* Access Service */
	accessService, err := service.NewAccessService(userService, db.Collection("apiKeys"))
	if err != nil {
		panic(err)
	}
	pack.AccessService = accessService
	sw.Lap("Init access service")

	/* Baseic global caster */
	caster := models.NewSimpleCaster(clientService)
	pack.Caster = caster

	/* Hasher */
	hasher := models.NewHasher(workerPool, caster)

	/* FileTree Service */
	mediaFileTree, err := fileTree.NewFileTree(
		env.GetMediaRoot(), "MEDIA", hasher,
		mediaJournal,
	)
	if err != nil {
		panic(err)
	}

	hollowJournal := mock.NewHollowJournalService()
	hollowHasher := mock.NewMockHasher()
	cachesTree, err := fileTree.NewFileTree(env.GetCachesRoot(), "CACHES", hollowHasher, hollowJournal)
	if err != nil {
		panic(err)
	}
	sw.Lap("Init file trees")

	fileService, err := service.NewFileService(
		mediaFileTree, cachesTree, userService, accessService, nil,
		db.Collection("trash"),
	)
	if err != nil {
		panic(err)
	}
	pack.FileService = fileService
	sw.Lap("Init file service")

	/* Media type Service */
	// Only from config file, for now
	marshMap := map[string]models.MediaType{}
	err = env.ReadTypesConfig(&marshMap)
	if err != nil {
		panic(err)
	}

	mediaTypeServ := models.NewTypeService(marshMap)
	/* Media Service */
	mediaService, err := service.NewMediaService(
		fileService, mediaTypeServ, &mock.MockAlbumService{},
		db.Collection("media"),
	)
	if err != nil {
		panic(err)
	}
	pack.MediaService = mediaService
	sw.Lap("Init media service")

	fileService.SetMediaService(mediaService)
	clientService.SetFileService(fileService)

	/* Album Service */
	albumService := service.NewAlbumService(db.Collection("albums"), mediaService, shareService)
	err = albumService.Init()
	if err != nil {
		panic(err)
	}
	pack.AlbumService = albumService
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

	log.Trace.Println("Service setup complete")
	pack.Loaded.Store(true)
}

func mainRecovery(msg string) {
	err := recover()
	if err != nil {
		log.ErrTrace(err.(error))
		log.ErrorCatcher.Println(msg)
		os.Exit(1)
	}
}
