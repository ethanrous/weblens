package main

import (
	"errors"
	_ "net/http/pprof"
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
	"github.com/ethrousseau/weblens/service/proxy"
	"github.com/ethrousseau/weblens/task"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var server *Server
var services = &models.ServicePack{}

func main() {
	config, err := env.ReadConfig(env.GetConfigName())
	if err != nil {
		panic(err)
	}

	defer mainRecovery("WEBLENS ENCOUNTERED AN UNRECOVERABLE ERROR")
	log.Info.Println("Starting Weblens")

	logLevel := env.GetLogLevel()
	log.SetLogLevel(logLevel)

	// if logLevel > 0 {
	// 	log.Debug.Println("Starting Weblens in debug mode")
	//
	// 	metricsServer := http.Server{
	// 		Addr:     "localhost:2112",
	// 		ErrorLog: log.ErrorCatcher,
	// 		Handler:  promhttp.Handler(),
	// 	}
	// 	go func() { log.ErrTrace(metricsServer.ListenAndServe()) }()
	// } else {
	// }
	gin.SetMode(gin.ReleaseMode)

	configName := env.GetConfigName()

	server = NewServer(config["routerHost"].(string), env.GetRouterPort(configName), services)
	server.StartupFunc = func() {
		startup(env.GetConfigName(), services, server)
	}
	services.StartupChan = make(chan bool)
	server.Start()
}

func startup(configName string, pack *models.ServicePack, srv *Server) {
	defer mainRecovery("WEBLENS STARTUP FAILED")

	log.Trace.Println("Beginning service setup")

	sw := internal.NewStopwatch("Initialization")


	/* Database connection */
	db, err := database.ConnectToMongo(env.GetMongoURI(configName), env.GetMongoDBName(configName))
	if err != nil {
		panic(err)
	}
	sw.Lap("Connect to Mongo")

	setupInstanceService(pack, db)
	sw.Lap("Init instance service")

	setupUserService(pack, db)
	sw.Lap("Init user service")

	setupTaskService(pack)
	sw.Lap("Worker pool enabled")

	/* Client Manager */
	pack.ClientService = service.NewClientManager(pack)
	sw.Lap("Init client service")

	/* Basic global pack.Caster */
	pack.Caster = models.NewSimpleCaster(pack.ClientService)

	if pack.InstanceService.GetLocal().GetRole() == models.InitServer {
		srv.UseInit()
	} else {
		srv.UseApi()
	}

	/* If server is backup server, connect to core server and launch backup daemon */
	if pack.InstanceService.GetLocal().GetRole() == models.BackupServer {
		core := pack.InstanceService.GetCore()
		if core == nil {
			panic(werror.Errorf("Could not find core instance"))
		}

		err = WebsocketToCore(core, services)
		if err != nil {
			panic(err)
		}

		var coreAddr string
		coreAddr, err = core.GetAddress()
		if err != nil {
			panic(err)
		}

		if coreAddr == "" || pack.InstanceService.GetCore().GetUsingKey() == "" {
			panic(werror.Errorf("could not get core address or key"))
		}

	} else if pack.InstanceService.GetLocal().GetRole() == models.CoreServer {
		srv.UseCore()
	}

	/* Share Service */
	pack.AddStartupTask("share_service", "Shares Service")
	shareService, err := service.NewShareService(db.Collection("shares"))
	if err != nil {
		panic(err)
	}
	pack.ShareService = shareService
	sw.Lap("Init share service")
	if err = pack.RemoveStartupTask("share_service"); err != nil {
		panic(err)
	}

	/* Journal Service */
	mediaJournal, err := fileTree.NewJournal(
		db.Collection("fileHistory"), pack.InstanceService.GetLocal().ServerId(),
		pack.InstanceService.GetLocal().GetRole() == models.BackupServer,
	)
	if err != nil {
		panic(err)
	}
	sw.Lap("Init journal service")

	/* Access Service */
	accessService, err := service.NewAccessService(pack.UserService, db.Collection("apiKeys"))
	if err != nil {
		panic(err)
	}
	pack.AccessService = accessService
	sw.Lap("Init access service")

	/* Hasher */
	hasher := models.NewHasher(pack.TaskService, pack.Caster)

	pack.AddStartupTask("file_services", "File Services")
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
		mediaFileTree, cachesTree, pack.UserService, accessService, nil,
		db.Collection("trash"),
	)
	if err != nil {
		panic(err)
	}
	sw.Lap("Init file service")
	if err = pack.RemoveStartupTask("file_services"); err != nil {
		panic(err)
	}

	if pack.InstanceService.GetLocal().GetRole() == models.CoreServer {
		event := fileService.GetMediaJournal().NewEvent()

		users, err := pack.UserService.GetAll()
		if err != nil {
			panic(err)
		}

		for u := range users {
			if u.IsSystemUser() {
				continue
			}

			home, err := fileService.CreateFolder(fileService.GetMediaRoot(), u.GetUsername(), pack.Caster)
			if err != nil && !errors.Is(err, werror.ErrDirAlreadyExists) {
				panic(err)
			}
			u.SetHomeFolder(home)

			trash, err := fileService.CreateFolder(home, ".user_trash", pack.Caster)
			if err != nil && !errors.Is(err, werror.ErrDirAlreadyExists) {
				panic(err)
			}
			u.SetTrashFolder(trash)
		}

		fileService.GetMediaJournal().LogEvent(event)
		sw.Lap("Find or create user directories")
	}

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
	pack.FileService = fileService

	/* Album Service */
	albumService := service.NewAlbumService(db.Collection("albums"), mediaService, shareService)
	err = albumService.Init()
	if err != nil {
		panic(err)
	}
	pack.AlbumService = albumService
	sw.Lap("Init album service")

	mediaService.AlbumService = albumService

	if pack.InstanceService.GetLocal().GetRole() == models.BackupServer {
		pack.AddStartupTask("core_connect", "Waiting for Core connection")
		for pack.ClientService.GetClientByServerId(pack.InstanceService.GetCore().ServerId()) == nil {
			time.Sleep(1 * time.Second)
		}
		err = pack.RemoveStartupTask("core_connect")
		if err != nil {
			panic(err)
		}

		go jobs.BackupD(time.Hour, pack)
	}

	sw.Stop()
	sw.PrintResults(false)
	log.Info.Printf(
		"Weblens loaded in %s. %d files, %d medias, and %d users\n", sw.GetTotalTime(false), fileService.Size(),
		mediaService.Size(), pack.UserService.Size(),
	)

	pack.Loaded.Store(true)
	close(pack.StartupChan)
	pack.StartupChan = nil

	log.Trace.Println("Service setup complete")
	pack.Caster.PushWeblensEvent("weblens_loaded")
}

func mainRecovery(msg string) {
	err := recover()
	if err != nil {
		log.ErrTrace(err.(error))
		log.ErrorCatcher.Println(msg)
		os.Exit(1)
	}
}

func setupInstanceService(pack *models.ServicePack, db *mongo.Database) {
	instanceService, err := service.NewInstanceService(db.Collection("servers"))
	if err != nil {
		panic(err)
	}
	pack.InstanceService = instanceService

	// Let router know it can start. Instance server is required for the most basic routes,
	// so we can start the router only after that's set.
	pack.StartupChan <- true

	log.Debug.Printf("Local server role is %s", pack.InstanceService.GetLocal().GetRole())
}

func setupUserService(pack *models.ServicePack, db *mongo.Database) {
	if pack.InstanceService.GetLocal().GetRole() == models.BackupServer {
		pack.UserService = proxy.NewProxyUserService(pack.InstanceService.GetCore())
	} else {
		userService, err := service.NewUserService(db.Collection("users"))
		if err != nil {
			panic(err)
		}
		pack.UserService = userService
	}
}

func setupTaskService(pack *models.ServicePack) {
	/* Enable the worker pool held by the task tracker
	loading the filesystem might dispatch tasks,
	so we have to start the pool first */
	workerPool := task.NewWorkerPool(env.GetWorkerCount(), env.GetLogLevel())

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
}
