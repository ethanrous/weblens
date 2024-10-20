package main

import (
	"errors"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"time"

	"github.com/ethanrous/weblens/database"
	"github.com/ethanrous/weblens/fileTree"
	. "github.com/ethanrous/weblens/http"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/jobs"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/service"
	"github.com/ethanrous/weblens/service/mock"
	"github.com/ethanrous/weblens/task"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	var server *Server
	var services = &models.ServicePack{}

	config, err := env.ReadConfig(env.GetConfigName())
	if err != nil {
		panic(err)
	}

	defer mainRecovery("WEBLENS ENCOUNTERED AN UNRECOVERABLE ERROR")
	log.Info.Println("Starting Weblens")

	configName := env.GetConfigName()
	log.Info.Printf("Using config: %s", configName)

	logLevel := env.GetLogLevel(configName)
	log.SetLogLevel(logLevel)

	if logLevel != log.TRACE {
		gin.SetMode(gin.ReleaseMode)
	}

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
	pack.Db = db
	sw.Lap("Connect to Mongo")

	setupInstanceService(pack)
	sw.Lap("Init instance service")

	localRole := pack.InstanceService.GetLocal().Role

	setupUserService(pack)
	sw.Lap("Init user service")

	setupTaskService(pack)
	sw.Lap("Worker pool enabled")

	/* Client Manager */
	pack.ClientService = service.NewClientManager(pack)
	sw.Lap("Init client service")

	/* Basic global pack.Caster */
	pack.Caster = models.NewSimpleCaster(pack.ClientService)

	setupAccessService(pack, db)
	sw.Lap("Init access service")

	/* Share Service */
	pack.AddStartupTask("share_service", "Shares Service")
	shareService, err := service.NewShareService(db.Collection("shares"))
	if err != nil {
		panic(err)
	}
	pack.ShareService = shareService
	sw.Lap("Init share service")
	pack.RemoveStartupTask("share_service")

	setupFileService(pack)
	sw.Lap("Init file service")

	// Add basic routes to the router
	if localRole == models.InitServer {
		// Uninitialized servers get "Init" routes
		srv.UseInit()
	} else {
		// All initialized servers get the "API" group of routes
		srv.UseApi()

		// If server is CORE, add core routes and discover user directories
		if localRole == models.CoreServer {
			srv.UseInterserverRoutes()

			event := pack.FileService.GetJournalByTree("USERS").NewEvent()

			users, err := pack.UserService.GetAll()
			if err != nil {
				panic(err)
			}

			for u := range users {
				if u.IsSystemUser() {
					continue
				}

				var hadNoHome bool
				if u.HomeId == "" {
					hadNoHome = true
				}

				err = pack.FileService.CreateUserHome(u)
				if err != nil {
					panic(err)
				}

				if hadNoHome {
					err = pack.UserService.UpdateUserHome(u)
					if err != nil {
						panic(err)
					}
				}
			}

			pack.FileService.GetJournalByTree("USERS").LogEvent(event)
			sw.Lap("Find or create user directories")
		} else if localRole == models.BackupServer {
			/* If server is backup server, connect to core server and launch backup daemon */
			pack.AddStartupTask("core_connect", "Waiting for Core connection")
			cores := pack.InstanceService.GetCores()
			if len(cores) == 0 {
				log.Error.Println("No core servers found in database")
			}

			for _, core := range cores {
				if core != nil {
					log.Trace.Printf("Connecting to core server [%s]", core.Address)
					if err = WebsocketToCore(core, pack); err != nil {
						panic(err)
					}

					var coreAddr string
					coreAddr, err = core.GetAddress()
					if err != nil {
						panic(err)
					}

					if coreAddr == "" || core.GetUsingKey() == "" {
						panic(werror.Errorf("could not get core address or key"))
					}
				}
			}
			pack.RemoveStartupTask("core_connect")

			go jobs.BackupD(time.Hour, pack)

		}

		setupMediaService(pack, db)
		sw.Lap("Init media service")

		setupAlbumService(pack, db)
		sw.Lap("Init album service")

		log.Info.Printf(
			"Weblens loaded in %s. %dB in files, %d medias, and %d users\n", sw.GetTotalTime(false),
			pack.FileService.(*service.FileServiceImpl).Size("USERS"),
			pack.MediaService.Size(), pack.UserService.Size(),
		)
	}

	sw.Stop()
	sw.PrintResults(false)

	pack.Loaded.Store(true)
	close(pack.StartupChan)
	pack.StartupChan = nil

	log.Trace.Println("Service setup complete")
	pack.Caster.PushWeblensEvent(models.WeblensLoadedEvent, models.WsC{"role": pack.InstanceService.GetLocal().GetRole()})
}

func mainRecovery(msg string) {
	err := recover()
	if err != nil {
		if _, ok := err.(error); !ok {
			err = werror.Errorf("%s", err)
		}
		log.ErrTrace(err.(error))
		log.ErrorCatcher.Println(msg)
		os.Exit(1)
	}
}

func setupInstanceService(pack *models.ServicePack) {
	instanceService, err := service.NewInstanceService(pack.Db.Collection("servers"))
	if err != nil {
		panic(err)
	}
	pack.InstanceService = instanceService

	// Let router know it can start. Instance server is required for the most basic routes,
	// so we can start the router only after that's set.
	pack.StartupChan <- true

	log.Debug.Printf("Local server role is %s", pack.InstanceService.GetLocal().Role)
}

func setupUserService(pack *models.ServicePack) {
	userService, err := service.NewUserService(pack.Db.Collection("users"))
	if err != nil {
		panic(err)
	}
	pack.UserService = userService
}

func setupTaskService(pack *models.ServicePack) {
	/* Enable the worker pool held by the task tracker
	loading the filesystem might dispatch tasks,
	so we have to start the pool first */
	workerPool := task.NewWorkerPool(env.GetWorkerCount(), log.GetLogLevel())

	workerPool.RegisterJob(models.ScanDirectoryTask, jobs.ScanDirectory)
	workerPool.RegisterJob(models.ScanFileTask, jobs.ScanFile)
	workerPool.RegisterJob(models.UploadFilesTask, jobs.HandleFileUploads)
	workerPool.RegisterJob(models.CreateZipTask, jobs.CreateZip)
	workerPool.RegisterJob(models.GatherFsStatsTask, jobs.GatherFilesystemStats)
	if pack.InstanceService.GetLocal().Role == models.BackupServer {
		workerPool.RegisterJob(models.BackupTask, jobs.DoBackup)
		workerPool.RegisterJob(models.CopyFileFromCoreTask, jobs.CopyFileFromCore)
		workerPool.RegisterJob(models.RestoreCoreTask, jobs.RestoreCore)
	} else if pack.InstanceService.GetLocal().Role == models.CoreServer {
		workerPool.RegisterJob(models.HashFileTask, jobs.HashFile)
	}

	pack.TaskService = workerPool
	workerPool.Run()
}

func setupFileService(pack *models.ServicePack) {
	pack.AddStartupTask("file_services", "File Services")

	/* Hasher */
	hasherFactory := func() fileTree.Hasher {
		return models.NewHasher(pack.TaskService, pack.Caster)
	}

	localRole := pack.InstanceService.GetLocal().Role

	/* Journal Service */
	var ignoreLocal bool
	if localRole == models.BackupServer || localRole == models.InitServer {
		ignoreLocal = true
	}
	mediaJournal, err := fileTree.NewJournal(
		pack.Db.Collection("fileHistory"), pack.InstanceService.GetLocal().ServerId(),
		ignoreLocal, hasherFactory,
	)
	if err != nil {
		panic(err)
	}

	var trees []fileTree.FileTree
	hollowJournal := mock.NewHollowJournalService()

	/* Restore FileTree */
	restoreFileTree, err := fileTree.NewFileTree(
		filepath.Join(env.GetDataRoot(), ".restore"), "RESTORE", hollowJournal, !ignoreLocal,
	)
	if err != nil {
		panic(err)
	}

	if localRole == models.CoreServer {
		/* Users FileTree */
		usersFileTree, err := fileTree.NewFileTree(
			filepath.Join(env.GetDataRoot(), "users"), "USERS", mediaJournal, !ignoreLocal,
		)
		if err != nil {
			panic(err)
		}

		cachesTree, err := fileTree.NewFileTree(env.GetCachesRoot(), "CACHES", hollowJournal, !ignoreLocal)
		if err != nil {
			panic(err)
		}
		_, err = cachesTree.MkDir(cachesTree.GetRoot(), "takeout", &fileTree.FileEvent{})
		if err != nil && !errors.Is(err, werror.ErrDirAlreadyExists) {
			panic(err)
		}

		trees = []fileTree.FileTree{usersFileTree, cachesTree, restoreFileTree}
	} else if localRole == models.BackupServer {
		for _, core := range pack.InstanceService.GetCores() {
			newJournal, err := fileTree.NewJournal(
				pack.Db.Collection("fileHistory"), core.ServerId(), true, hasherFactory,
			)
			if err != nil {
				panic(err)
			}

			newTree, err := fileTree.NewFileTree(
				filepath.Join(env.GetDataRoot(), core.ServerId()), core.ServerId(), newJournal, false,
			)
			if err != nil {
				panic(err)
			}

			trees = append(trees, newTree)
		}
		trees = append(trees, restoreFileTree)
	}

	fileService, err := service.NewFileService(
		pack.InstanceService,
		pack.UserService,
		pack.AccessService,
		nil,
		pack.Db.Collection("trash"),
		pack.Db.Collection("folderMedia"),
		trees...,
	)
	if err != nil {
		panic(err)
	}

	pack.RemoveStartupTask("file_services")

	pack.FileService = fileService

	for _, tree := range trees {
		err = pack.FileService.ResizeDown(tree.GetRoot(), pack.Caster)
		if err != nil {
			panic(err)
		}
	}
}

func setupMediaService(pack *models.ServicePack, db *mongo.Database) {
	/* Media type Service */
	// Only from config file, for now
	marshMap := map[string]models.MediaType{}
	err := env.ReadTypesConfig(&marshMap)
	if err != nil {
		panic(err)
	}

	mediaTypeServ := models.NewTypeService(marshMap)
	/* Media Service */
	mediaService, err := service.NewMediaService(
		pack.FileService, mediaTypeServ, &mock.MockAlbumService{},
		db.Collection("media"),
	)
	if err != nil {
		panic(err)
	}
	pack.MediaService = mediaService

	pack.FileService.(*service.FileServiceImpl).SetMediaService(mediaService)
}

func setupAlbumService(pack *models.ServicePack, db *mongo.Database) {
	/* Album Service */
	albumService := service.NewAlbumService(db.Collection("albums"), pack.MediaService, pack.ShareService)
	err := albumService.Init()
	if err != nil {
		panic(err)
	}
	pack.AlbumService = albumService

	pack.MediaService.(*service.MediaServiceImpl).AlbumService = albumService
}

func setupAccessService(pack *models.ServicePack, db *mongo.Database) {
	/* Access Service */
	accessService, err := service.NewAccessService(pack.UserService, db.Collection("apiKeys"))
	if err != nil {
		panic(err)
	}
	pack.AccessService = accessService

	keys, err := accessService.GetAllKeys(pack.UserService.GetRootUser())
	if err != nil {
		panic(err)
	}
	for _, key := range keys {
		if key.RemoteUsing != "" {
			// Connect to remote server
			remote := pack.InstanceService.GetByInstanceId(key.RemoteUsing)
			if remote != nil {
				continue
			}

			err = accessService.SetKeyUsedBy(key.Key, nil)
			if err != nil {
				panic(err)
			}
		}
	}
}
