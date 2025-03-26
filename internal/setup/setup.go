package setup

// import (
// 	"os"
// 	"path/filepath"
// 	"time"
//
// 	"github.com/ethanrous/weblens/database"
// 	"github.com/ethanrous/weblens/fileTree"
// 	"github.com/ethanrous/weblens/http"
// 	"github.com/ethanrous/weblens/internal"
// 	"github.com/ethanrous/weblens/internal/env"
// 	"github.com/ethanrous/weblens/internal/werror"
// 	"github.com/ethanrous/weblens/jobs"
// 	"github.com/ethanrous/weblens/models"
// 	"github.com/ethanrous/weblens/models/caster"
// 	"github.com/ethanrous/weblens/service"
// 	"github.com/ethanrous/weblens/service/mock"
// 	"github.com/ethanrous/weblens/services"
// 	"github.com/ethanrous/weblens/services/context"
// 	"github.com/ethanrous/weblens/task"
// 	"github.com/pkg/errors"
// 	"github.com/rs/zerolog"
// 	"go.mongodb.org/mongo-driver/mongo"
// )
//
// func MainRecovery(logger *zerolog.Logger) {
// 	err := recover()
// 	if err != nil {
// 		if erri, ok := err.(error); ok {
// 			erri = errors.Wrap(errors.New(erri.Error()), "Panic")
// 			logger.Fatal().Stack().Err(erri).Msgf("Unrecoverable error occurred during router startup")
// 		} else {
// 			err := errors.Wrap(errors.New("Panic"), "Panic")
// 			logger.Fatal().Stack().Err(err).Msgf("Unrecoverable error occurred during router startup: %v", err)
// 		}
// 		os.Exit(1)
// 	}
// }
//
// func Startup(cnf env.Config, ctx *context.BasicContext) {
// 	defer MainRecovery(pack.Log)
// 	start := time.Now()
//
// 	pack.Log.Trace().Msg("Service setup started")
//
// 	/* Database connection */
// 	db, err := database.ConnectToMongo(cnf.MongodbUri, cnf.MongodbName, pack.Log)
// 	if err != nil {
// 		panic(err)
// 	}
// 	pack.Db = db
// 	pack.Log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Connected to mongodb") })
//
// 	setupInstanceService(pack)
//
// 	localRole := pack.InstanceService.GetLocal().Role
//
// 	setupUserService(pack)
//
// 	setupTaskService(cnf.WorkerCount, pack, pack.Log)
//
// 	/* Client Manager */
// 	pack.ClientService = service.NewClientManager(pack, pack.Log)
//
// 	/* Basic global pack.Caster */
// 	caster := caster.NewSimpleCaster(pack.ClientService, pack.Log)
// 	caster.Global()
// 	pack.SetCaster(caster)
//
// 	setupAccessService(pack, db)
//
// 	setupShareService(pack, db)
//
// 	setupFileService(cnf.DataRoot, cnf.CachesRoot, pack)
//
// 	// Add basic routes to the router
// 	if localRole != models.InitServerRole {
// 		// If server is CORE, add core routes and discover user directories
// 		if localRole == models.CoreServerRole {
// 			// srv.UseInterserverRoutes()
//
// 		} else if localRole == models.BackupServerRole {
// 			/* If server is backup server, connect to core server and launch backup daemon */
// 			pack.AddStartupTask("core_connect", "Waiting for Core connection")
// 			cores := pack.InstanceService.GetCores()
// 			if len(cores) == 0 {
// 				pack.Log.Error().Msg("No core servers found in database")
// 			}
//
// 			for _, core := range cores {
// 				if core != nil {
// 					pack.Log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Connecting to core server [%s]", core.Address) })
// 					if err = http.WebsocketToCore(core, pack); err != nil {
// 						panic(err)
// 					}
//
// 					var coreAddr string
// 					coreAddr, err = core.GetAddress()
// 					if err != nil {
// 						panic(err)
// 					}
//
// 					if coreAddr == "" || core.GetUsingKey() == "" {
// 						panic(werror.Errorf("could not get core address or key"))
// 					}
// 				}
// 			}
// 			pack.RemoveStartupTask("core_connect")
//
// 			go jobs.BackupD(time.Hour, pack)
//
// 		}
//
// 		setupMediaService(pack, db)
//
// 		setupAlbumService(pack, db)
//
// 		pack.Log.Info().Msgf(
// 			"Weblens loaded in %s. %s in files, %d medias, and %d users", time.Since(start),
// 			internal.ByteCountSI(pack.FileService.(*service.FileServiceImpl).Size("USERS")),
// 			pack.MediaService.Size(), pack.UserService.Size(),
// 		)
// 	}
//
// 	server := pack.Server.(*http.Server)
// 	server.RouterLock.Lock()
// 	pack.Loaded.Store(true)
//
// 	pack.Log.Trace().Msg("Closing startup channel")
//
// 	close(pack.StartupChan)
// 	pack.StartupChan = nil
// 	server.RouterLock.Unlock()
//
// 	pack.Log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Service setup complete") })
// 	pack.GetCaster().PushWeblensEvent(models.WeblensLoadedEvent, models.WsC{"role": pack.InstanceService.GetLocal().GetRole()})
// }
//
// func setupInstanceService(ctx *context.BasicContext) {
// 	instanceService, err := service.NewInstanceService(pack.Db.Collection(string(database.InstanceCollectionKey)), pack.Log)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	pack.InstanceService = instanceService
// 	pack.Log.UpdateContext(func(c zerolog.Context) zerolog.Context {
// 		return c.Str("instance", pack.InstanceService.GetLocal().ServerId())
// 	})
//
// 	// Let router know it can start. Instance server is required for the most basic routes,
// 	// so we can start the router only after that's set.
// 	pack.StartupChan <- true
//
// 	pack.Log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Local server role is %s", pack.InstanceService.GetLocal().Role) })
// }
//
// func setupUserService(ctx *context.BasicContext) {
// 	userService, err := service.NewUserService(pack.Db.Collection(string(database.UsersCollectionKey)))
// 	if err != nil {
// 		panic(err)
// 	}
// 	pack.UserService = userService
// }
//
// func setupTaskService(workerCount int, ctx *context.BasicContext, logger *zerolog.Logger) {
// 	// Loading the filesystem might dispatch tasks,
// 	// so we have to start the pool first
// 	workerPool := task.NewWorkerPool(workerCount, logger)
//
// 	jobs.RegisterJobs(workerPool, pack.InstanceService.GetLocal().Role)
//
// 	pack.TaskService = workerPool
// 	workerPool.Run()
// }
//
// func setupFileService(dataRootPath, cachesRootPath string, ctx *context.BasicContext) {
// 	pack.AddStartupTask("file_services", "Setting up File Services")
//
// 	/* Hasher */
// 	hasherFactory := func() fileTree.Hasher {
// 		return models.NewHasher(pack.TaskService, pack.GetCaster())
// 	}
//
// 	localRole := pack.InstanceService.GetLocal().Role
//
// 	/* Journal Service */
// 	journalConfig := fileTree.JournalConfig{
// 		Collection:    pack.Db.Collection(string(database.FileHistoryCollectionKey)),
// 		ServerId:      pack.InstanceService.GetLocal().ServerId(),
// 		HasherFactory: hasherFactory,
// 		Logger:        pack.Log,
// 	}
//
// 	if localRole == models.BackupServerRole || localRole == models.InitServerRole {
// 		journalConfig.IgnoreLocal = true
// 	}
//
// 	mediaJournal, err := fileTree.NewJournal(journalConfig)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	var trees []fileTree.FileTree
// 	hollowJournal := mock.NewHollowJournalService()
//
// 	/* Restore FileTree */
// 	restoreFileTree, err := fileTree.NewFileTree(
// 		filepath.Join(dataRootPath, ".restore"), "RESTORE", hollowJournal, !journalConfig.IgnoreLocal, pack.Log,
// 	)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	if localRole == models.CoreServerRole {
// 		/* Users FileTree */
// 		usersFileTree, err := fileTree.NewFileTree(
// 			filepath.Join(dataRootPath, "users"), "USERS", mediaJournal, !journalConfig.IgnoreLocal, pack.Log,
// 		)
// 		if err != nil {
// 			panic(err)
// 		}
//
// 		cachesTree, err := fileTree.NewFileTree(cachesRootPath, "CACHES", hollowJournal, !journalConfig.IgnoreLocal, pack.Log)
// 		if err != nil {
// 			panic(err)
// 		}
// 		_, err = cachesTree.MkDir(cachesTree.GetRoot(), "takeout", &fileTree.FileEvent{})
// 		if err != nil && !errors.Is(err, werror.ErrDirAlreadyExists) {
// 			panic(err)
// 		}
// 		_, err = cachesTree.MkDir(cachesTree.GetRoot(), "thumbs", &fileTree.FileEvent{})
// 		if err != nil && !errors.Is(err, werror.ErrDirAlreadyExists) {
// 			panic(err)
// 		}
//
// 		trees = []fileTree.FileTree{usersFileTree, cachesTree, restoreFileTree}
//
// 	} else if localRole == models.BackupServerRole {
// 		for _, core := range pack.InstanceService.GetCores() {
// 			journalConfig := fileTree.JournalConfig{
// 				IgnoreLocal:   true,
// 				Collection:    pack.Db.Collection(string(database.FileHistoryCollectionKey)),
// 				ServerId:      core.ServerId(),
// 				HasherFactory: hasherFactory,
// 				Logger:        pack.Log,
// 			}
//
// 			newJournal, err := fileTree.NewJournal(journalConfig)
// 			if err != nil {
// 				panic(err)
// 			}
//
// 			newTree, err := fileTree.NewFileTree(
// 				filepath.Join(dataRootPath, core.ServerId()), core.ServerId(), newJournal, false, pack.Log,
// 			)
// 			if err != nil {
// 				panic(err)
// 			}
//
// 			trees = append(trees, newTree)
// 		}
// 		if len(trees) == 0 {
// 			panic(werror.Errorf("No core servers found in database"))
// 		}
//
// 		trees = append(trees, restoreFileTree)
// 	}
//
// 	treeNames := make([]string, len(trees))
// 	for i, tree := range trees {
// 		treeNames[i] = tree.GetRoot().Filename()
// 	}
// 	pack.Log.Debug().Msgf("File trees: %v", treeNames)
//
// 	fileService, err := service.NewFileService(
// 		pack.Log,
// 		pack.InstanceService,
// 		pack.UserService,
// 		pack.AccessService,
// 		nil,
// 		pack.Db.Collection(string(database.FolderMediaCollectionKey)),
// 		trees...,
// 	)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	pack.SetFileService(fileService)
//
// 	for _, tree := range trees {
// 		err = pack.FileService.ResizeDown(tree.GetRoot(), nil, pack.GetCaster())
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
//
// 	if localRole == models.CoreServerRole {
// 		event := pack.FileService.GetJournalByTree("USERS").NewEvent()
// 		users, err := pack.UserService.GetAll()
// 		if err != nil {
// 			panic(err)
// 		}
//
// 		for u := range users {
// 			if u.IsSystemUser() {
// 				continue
// 			}
//
// 			var hadNoHome bool
// 			if u.HomeId == "" {
// 				hadNoHome = true
// 			}
//
// 			err = pack.FileService.CreateUserHome(u)
// 			if err != nil {
// 				panic(err)
// 			}
//
// 			if hadNoHome {
// 				err = pack.UserService.UpdateUserHome(u)
// 				if err != nil {
// 					panic(err)
// 				}
// 			}
// 		}
//
// 		pack.FileService.GetJournalByTree("USERS").LogEvent(event)
// 	}
// 	pack.RemoveStartupTask("file_services")
// }
//
// func setupMediaService(ctx *context.BasicContext) {
// 	pack.AddStartupTask("media_service", "Setting up Media Service")
// 	// Only from config file, for now
// 	marshMap := map[string]models.MediaType{}
// 	err := env.ReadTypesConfig(&marshMap)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	mediaTypeServ := models.NewTypeService(marshMap)
// 	mediaService, err := service.NewMediaService(
// 		pack.FileService, mediaTypeServ, &mock.MockAlbumService{},
// 		pack.Db.Collection(string(database.MediaCollectionKey)), pack.Log,
// 	)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	pack.MediaService = mediaService
// 	pack.FileService.(*service.FileServiceImpl).SetMediaService(mediaService)
//
// 	pack.RemoveStartupTask("media_service")
// }
//
// func setupAlbumService(ctx *context.BasicContext, db *mongo.Database) {
// 	pack.AddStartupTask("album_service", "Setting up Album Service")
//
// 	albumService := service.NewAlbumService(db.Collection(string(database.AlbumsCollectionKey)), pack.MediaService, pack.ShareService)
// 	err := albumService.Init()
// 	if err != nil {
// 		panic(err)
// 	}
// 	pack.AlbumService = albumService
// 	pack.MediaService.(*service.MediaServiceImpl).AlbumService = albumService
//
// 	pack.RemoveStartupTask("album_service")
// }
//
// func setupAccessService(ctx *context.BasicContext, db *mongo.Database) {
// 	pack.AddStartupTask("access_service", "Setting up Access Service")
//
// 	accessService, err := service.NewAccessService(pack.UserService, db.Collection(string(database.ApiKeysCollectionKey)))
// 	if err != nil {
// 		panic(err)
// 	}
// 	pack.AccessService = accessService
//
// 	keys, err := accessService.GetKeysByUser(pack.UserService.GetRootUser())
// 	if err != nil {
// 		panic(err)
// 	}
// 	for _, key := range keys {
// 		if key.RemoteUsing != "" {
// 			// Connect to remote server
// 			remote := pack.InstanceService.GetByInstanceId(key.RemoteUsing)
// 			if remote != nil {
// 				continue
// 			}
//
// 			err = accessService.SetKeyUsedBy(key.Key, nil)
// 			if err != nil {
// 				panic(err)
// 			}
// 		}
// 	}
// 	pack.RemoveStartupTask("access_service")
// }
//
// func setupShareService(ctx *context.BasicContext, db *mongo.Database) {
// 	pack.AddStartupTask("share_service", "Setting up Shares Service")
// 	shareService, err := service.NewShareService(db.Collection(string(database.SharesCollectionKey)), pack.Log)
// 	if err != nil {
// 		panic(err)
// 	}
// 	pack.ShareService = shareService
// 	pack.RemoveStartupTask("share_service")
// }
