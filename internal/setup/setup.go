package setup

import (
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ethanrous/weblens/database"
	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/http"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/jobs"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/caster"
	"github.com/ethanrous/weblens/service"
	"github.com/ethanrous/weblens/service/mock"
	"github.com/ethanrous/weblens/task"
	"go.mongodb.org/mongo-driver/mongo"
)

func MainRecovery(msg string, logger log.Bundle) {
	err := recover()
	if err != nil {
		if _, ok := err.(error); !ok {
			err = werror.Errorf("%s", err)
		}
		logger.ErrTrace(err.(error))
		logger.Error.Println(msg)
		os.Exit(1)
	}
}

func Startup(cnf env.Config, pack *models.ServicePack) {
	defer MainRecovery("WEBLENS STARTUP FAILED", pack.Log)

	pack.Log.Trace.Println("Beginning service setup")

	sw := internal.NewStopwatch("Initialization")

	/* Database connection */
	db, err := database.ConnectToMongo(cnf.MongodbUri, cnf.MongodbName)
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

	setupTaskService(cnf.WorkerCount, pack, pack.Log)
	sw.Lap("Worker pool enabled")

	/* Client Manager */
	pack.ClientService = service.NewClientManager(pack)
	sw.Lap("Init client service")

	/* Basic global pack.Caster */
	caster := caster.NewSimpleCaster(pack.ClientService)
	caster.Global()
	pack.SetCaster(caster)

	setupAccessService(pack, db)
	sw.Lap("Init access service")

	/* Share Service */
	pack.AddStartupTask("share_service", "Setting up Shares Service")
	shareService, err := service.NewShareService(db.Collection(string(database.SharesCollectionKey)))
	if err != nil {
		panic(err)
	}
	pack.ShareService = shareService
	sw.Lap("Init share service")
	pack.RemoveStartupTask("share_service")

	setupFileService(cnf.DataRoot, cnf.CachesRoot, pack)
	sw.Lap("Init file service")

	// Add basic routes to the router
	if localRole != models.InitServerRole {
		// If server is CORE, add core routes and discover user directories
		if localRole == models.CoreServerRole {
			// srv.UseInterserverRoutes()

			sw.Lap("Find or create user directories")
		} else if localRole == models.BackupServerRole {
			/* If server is backup server, connect to core server and launch backup daemon */
			pack.AddStartupTask("core_connect", "Waiting for Core connection")
			cores := pack.InstanceService.GetCores()
			if len(cores) == 0 {
				pack.Log.Error.Println("No core servers found in database")
			}

			for _, core := range cores {
				if core != nil {
					pack.Log.Trace.Func(func(l log.Logger) { l.Printf("Connecting to core server [%s]", core.Address) })
					if err = http.WebsocketToCore(core, pack); err != nil {
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

		pack.Log.Info.Printf(
			"Weblens loaded in %s. %s in files, %d medias, and %d users\n", sw.GetTotalTime(false),
			internal.ByteCountSI(pack.FileService.(*service.FileServiceImpl).Size("USERS")),
			pack.MediaService.Size(), pack.UserService.Size(),
		)
	}

	sw.Stop()
	sw.PrintResults(false)

	server := pack.Server.(*http.Server)
	server.RouterLock.Lock()
	pack.Loaded.Store(true)

	pack.Log.Trace.Println("Closing startup channel")

	close(pack.StartupChan)
	pack.StartupChan = nil
	server.RouterLock.Unlock()

	pack.Log.Debug.Println("Service setup complete")
	pack.GetCaster().PushWeblensEvent(models.WeblensLoadedEvent, models.WsC{"role": pack.InstanceService.GetLocal().GetRole()})

	// If we're in debug mode, wait for a SIGQUIT to exit,
	// this is to allow for easy restarting of the server
	if pack.Log.Level >= log.DEBUG {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGQUIT)
		<-sigs
		pack.Log.Info.Println("\nReceived SIGQUIT, exiting...")
		os.Exit(0)
	}
}

func setupInstanceService(pack *models.ServicePack) {
	instanceService, err := service.NewInstanceService(pack.Db.Collection(string(database.InstanceCollectionKey)))
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
	userService, err := service.NewUserService(pack.Db.Collection(string(database.UsersCollectionKey)))
	if err != nil {
		panic(err)
	}
	pack.UserService = userService
}

func setupTaskService(workerCount int, pack *models.ServicePack, logger log.Bundle) {
	// Loading the filesystem might dispatch tasks,
	// so we have to start the pool first
	workerPool := task.NewWorkerPool(workerCount, logger)

	workerPool.RegisterJob(models.ScanDirectoryTask, jobs.ScanDirectory)
	workerPool.RegisterJob(models.ScanFileTask, jobs.ScanFile)
	workerPool.RegisterJob(models.UploadFilesTask, jobs.HandleFileUploads, task.TaskOptions{Persistent: true, Unique: true})
	workerPool.RegisterJob(models.CreateZipTask, jobs.CreateZip)
	workerPool.RegisterJob(models.GatherFsStatsTask, jobs.GatherFilesystemStats)
	if pack.InstanceService.GetLocal().Role == models.BackupServerRole {
		workerPool.RegisterJob(models.BackupTask, jobs.DoBackup)
		workerPool.RegisterJob(models.CopyFileFromCoreTask, jobs.CopyFileFromCore)
		workerPool.RegisterJob(models.RestoreCoreTask, jobs.RestoreCore)
	} else if pack.InstanceService.GetLocal().Role == models.CoreServerRole {
		workerPool.RegisterJob(models.HashFileTask, jobs.HashFile)
	}

	pack.TaskService = workerPool
	workerPool.Run()
}

func setupFileService(dataRootPath, cachesRootPath string, pack *models.ServicePack) {
	pack.AddStartupTask("file_services", "Setting up File Services")

	sw := internal.NewStopwatch("Setup File Service")

	/* Hasher */
	hasherFactory := func() fileTree.Hasher {
		return models.NewHasher(pack.TaskService, pack.GetCaster())
	}
	sw.Lap("New Hasher")

	localRole := pack.InstanceService.GetLocal().Role

	/* Journal Service */
	var ignoreLocal bool
	if localRole == models.BackupServerRole || localRole == models.InitServerRole {
		ignoreLocal = true
	}
	mediaJournal, err := fileTree.NewJournal(pack.Db.Collection(string(database.FileHistoryCollectionKey)), pack.InstanceService.GetLocal().ServerId(), ignoreLocal, hasherFactory, pack.Log)
	if err != nil {
		panic(err)
	}
	sw.Lap("Init Media Journal")

	var trees []fileTree.FileTree
	hollowJournal := mock.NewHollowJournalService()

	/* Restore FileTree */
	restoreFileTree, err := fileTree.NewFileTree(
		filepath.Join(dataRootPath, ".restore"), "RESTORE", hollowJournal, !ignoreLocal,
	)
	if err != nil {
		panic(err)
	}
	sw.Lap("Init Restore Tree")

	if localRole == models.CoreServerRole {
		/* Users FileTree */
		usersFileTree, err := fileTree.NewFileTree(
			filepath.Join(dataRootPath, "users"), "USERS", mediaJournal, !ignoreLocal,
		)
		if err != nil {
			panic(err)
		}
		sw.Lap("Init Users Tree")

		cachesTree, err := fileTree.NewFileTree(cachesRootPath, "CACHES", hollowJournal, !ignoreLocal)
		if err != nil {
			panic(err)
		}
		sw.Lap("Init Caches Tree")
		_, err = cachesTree.MkDir(cachesTree.GetRoot(), "takeout", &fileTree.FileEvent{})
		if err != nil && !errors.Is(err, werror.ErrDirAlreadyExists) {
			panic(err)
		}
		sw.Lap("Init Caches Takeout Dir")
		_, err = cachesTree.MkDir(cachesTree.GetRoot(), "thumbs", &fileTree.FileEvent{})
		if err != nil && !errors.Is(err, werror.ErrDirAlreadyExists) {
			panic(err)
		}
		sw.Lap("Init Caches Thumbs Dir")

		trees = []fileTree.FileTree{usersFileTree, cachesTree, restoreFileTree}

	} else if localRole == models.BackupServerRole {
		for _, core := range pack.InstanceService.GetCores() {
			newJournal, err := fileTree.NewJournal(pack.Db.Collection(string(database.FileHistoryCollectionKey)), core.ServerId(), true, hasherFactory, pack.Log)
			if err != nil {
				panic(err)
			}

			newTree, err := fileTree.NewFileTree(
				filepath.Join(dataRootPath, core.ServerId()), core.ServerId(), newJournal, false,
			)
			if err != nil {
				panic(err)
			}

			trees = append(trees, newTree)
		}
		trees = append(trees, restoreFileTree)
	}

	fileService, err := service.NewFileService(
		pack.Log,
		pack.InstanceService,
		pack.UserService,
		pack.AccessService,
		nil,
		pack.Db.Collection(string(database.FolderMediaCollectionKey)),
		trees...,
	)
	if err != nil {
		panic(err)
	}
	sw.Lap("Create File Service")

	pack.SetFileService(fileService)

	for _, tree := range trees {
		err = pack.FileService.ResizeDown(tree.GetRoot(), nil, pack.GetCaster())
		if err != nil {
			panic(err)
		}
	}
	sw.Lap("Resize Trees")

	if localRole == models.CoreServerRole {
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
		sw.Lap("Verify User Directories")
	}
	sw.Stop()
	sw.PrintResults(false)

	pack.RemoveStartupTask("file_services")
}

func setupMediaService(pack *models.ServicePack, db *mongo.Database) {
	pack.AddStartupTask("media_service", "Setting up Media Service")
	// Only from config file, for now
	marshMap := map[string]models.MediaType{}
	err := env.ReadTypesConfig(&marshMap)
	if err != nil {
		panic(err)
	}

	mediaTypeServ := models.NewTypeService(marshMap)
	mediaService, err := service.NewMediaService(
		pack.FileService, mediaTypeServ, &mock.MockAlbumService{},
		db.Collection(string(database.MediaCollectionKey)), pack.Log,
	)
	if err != nil {
		panic(err)
	}

	pack.MediaService = mediaService
	pack.FileService.(*service.FileServiceImpl).SetMediaService(mediaService)

	pack.RemoveStartupTask("media_service")
}

func setupAlbumService(pack *models.ServicePack, db *mongo.Database) {
	pack.AddStartupTask("album_service", "Setting up Album Service")

	albumService := service.NewAlbumService(db.Collection(string(database.AlbumsCollectionKey)), pack.MediaService, pack.ShareService)
	err := albumService.Init()
	if err != nil {
		panic(err)
	}
	pack.AlbumService = albumService
	pack.MediaService.(*service.MediaServiceImpl).AlbumService = albumService

	pack.RemoveStartupTask("album_service")
}

func setupAccessService(pack *models.ServicePack, db *mongo.Database) {
	pack.AddStartupTask("access_service", "Setting up Access Service")

	accessService, err := service.NewAccessService(pack.UserService, db.Collection(string(database.ApiKeysCollectionKey)))
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
	pack.RemoveStartupTask("access_service")
}
