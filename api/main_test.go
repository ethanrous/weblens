package main

import (
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
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"runtime"
	"testing"
)

func TestInit(t *testing.T) {
	dbService := database.New("mongodb://localhost:27017", "weblens-test")
	types.SERV.SetDatabase(dbService)

	instanceService := instance.NewService()
	err := instanceService.Init(dbService)
	if err != nil {
		panic(err)
	}
	types.SERV.SetInstance(instanceService)

	ft := filetree.NewFileTree()
	err = ft.Init(dbService)
	if err != nil {
		panic(err)
	}
	types.SERV.SetFileTree(ft)

	userService := user.NewService()
	err = userService.Init(dbService)
	if err != nil {
		panic(err)
	}
	types.SERV.SetUserService(userService)

	instanceService.GetLocal().SetUserCount(types.SERV.UserService.Size())

	mediaTypeServ := media.NewTypeService()
	mediaService := media.NewRepo(mediaTypeServ)
	err = mediaService.Init(dbService)
	if err != nil {
		panic(err)
	}
	types.SERV.SetMediaRepo(mediaService)

	albumService := album.NewService()
	err = albumService.Init(dbService)
	if err != nil {
		panic(err)
	}
	types.SERV.SetAlbumService(albumService)

	clientService := routes.NewClientManager()
	types.SERV.SetClientService(clientService)

	workerPool, taskDispatcher := dataProcess.NewWorkerPool(runtime.NumCPU() - 2)
	types.SERV.SetWorkerPool(workerPool)
	types.SERV.SetTaskDispatcher(taskDispatcher)

	journal := history.NewService(ft, dbService)
	if journal == nil {
		panic("Cannot initialize journal")
	}
	err = journal.Init(dbService)
	if err != nil {
		panic(err)
	}

	ft.SetJournal(journal)
	go journal.JournalWorker()
	go journal.FileWatcher()

	requester := routes.NewRequester()

	localServer := instanceService.GetLocal()
	if localServer == nil {
		panic("Local server not initialized")
	}

	if localServer.ServerRole() == types.Backup {
		checkCoreExists(requester, nil)
	}

	caster := routes.NewBufferedCaster()
	types.SERV.SetCaster(caster)

	// Enable the worker pool held by the task tracker
	// loading the filesystem might dispatch tasks,
	// so we have to start the pool first
	workerPool.Run()

	// Load filesystem
	util.Info.Println("Loading filesystem...")
	dataStore.FsInit(ft)

	shareService := share.NewService()
	err = shareService.Init(dbService)
	if err != nil {
		panic(err)
	}
	types.SERV.SetShareService(shareService)

	err = dataStore.ClearTempDir(ft)
	util.FailOnError(err, "Failed to clear temporary directory on startup")

	err = dataStore.ClearTakeoutDir(ft)
	util.FailOnError(err, "Failed to clear takeout directory on startup")

	dataStore.InitApiKeyMap()

	// The global broadcaster is disabled by default so all the
	// initial loading of the filesystem (that was just done above) doesn't
	// try to broadcast for every file that exists. So it must be enabled here
	caster.DropBuffer()
}

func TestScan(t *testing.T) {

	filePath := "/Users/ethan/weblens/test/"

	//types.SERV.FileTree.NewFile()
	types.SERV.FileTree.GenerateFileId(filePath)

	//types.SERV.TaskDispatcher.ScanDirectory()
}
