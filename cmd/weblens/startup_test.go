package main

import (
	"context"
	"math/rand/v2"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethanrous/weblens/database"
	"github.com/ethanrous/weblens/http"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/setup"
	"github.com/ethanrous/weblens/internal/tests"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/jobs"
	"github.com/ethanrous/weblens/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func waitForStartup(startupChan chan bool) error {
	log.Debug.Println("Waiting for startup...")
	for {
		select {
		case sig, ok := <-startupChan:
			if ok {
				log.Debug.Println("Relaying startup signal")
				startupChan <- sig
			} else {
				return nil
			}
		case <-time.After(time.Second * 10):
			return werror.Errorf("Startup timeout")
		}
	}
}

func TestStartupCore(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping %s in short mode", t.Name())
	}

	t.Parallel()

	var server *http.Server

	cnf := env.Config{
		RouterHost:  env.GetRouterHost(),
		RouterPort:  rand.IntN(2000) + 8080,
		MongodbUri:  env.GetMongoURI(),
		MongodbName: "weblens-" + t.Name(),
		WorkerCount: 2,
		DataRoot:    filepath.Join(env.GetBuildDir(), "fs/test", t.Name(), "data"),
		CachesRoot:  filepath.Join(env.GetBuildDir(), "fs/test", t.Name(), "cache"),
		UiPath:      env.GetUIPath(),
	}

	var services = &models.ServicePack{
		Cnf: cnf,
		Log: log.NewLogPackage("", log.DEBUG),
	}

	mondb, err := database.ConnectToMongo(cnf.MongodbUri, cnf.MongodbName)
	require.NoError(t, err)

	err = mondb.Drop(context.Background())
	require.NoError(t, err)

	start := time.Now()
	server = http.NewServer(cnf.RouterHost, cnf.RouterPort, services)
	services.StartupChan = make(chan bool)
	server.StartupFunc = func() {
		setup.Startup(cnf, services)
	}
	go server.Start()

	if err := waitForStartup(services.StartupChan); err != nil {
		t.Fatal(err)
	}

	log.Debug.Println("Init startup took", time.Since(start))
	assert.True(t, services.Loaded.Load())

	err = services.InstanceService.InitCore("TEST-CORE")
	log.ErrTrace(err)
	require.NoError(t, err)

	// Although Restart() is safely synchronous outside of an HTTP request,
	// we call it without waiting to allow for our own timeout logic to be used
	services.Server.Restart(false)
	if err := waitForStartup(services.StartupChan); err != nil {
		t.Fatal(err)
	}
	log.Debug.Println("Core startup took", time.Since(start))
	assert.True(t, services.Loaded.Load())

	_, err = services.UserService.CreateOwner("test-username", "test-password", "Test Owner")
	if err != nil {
		t.Fatal(err)
	}

	services.Server.Restart(false)
	if err := waitForStartup(services.StartupChan); err != nil {
		t.Fatal(err)
	}
	log.Debug.Println("Core restart startup took", time.Since(start))
	assert.True(t, services.Loaded.Load())

	usersTree := services.FileService.GetFileTreeByName("USERS")
	if usersTree == nil {
		t.Fatal("No users tree")
	}

	_, err = usersTree.GetRoot().GetChild("test-username")
	assert.NoError(t, err)

	services.Server.Stop()
}

func TestStartupBackup(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping %s in short mode", t.Name())
	}

	t.Parallel()

	logger := log.NewLogPackage("", log.TRACE)

	coreServices, err := tests.NewWeblensTestInstance(t.Name(), env.Config{
		Role: string(models.CoreServerRole),
	})
	require.NoError(t, err)

	coreKeys, err := coreServices.AccessService.GetKeysByUser(coreServices.UserService.Get("test-username"))
	require.NoError(t, err)
	coreApiKey := coreKeys[0].Key
	coreAddress := env.GetProxyAddress(coreServices.Cnf)

	cnf := env.Config{
		RouterHost:  env.GetRouterHost(),
		RouterPort:  rand.IntN(2000) + 8080,
		MongodbUri:  env.GetMongoURI(),
		MongodbName: "weblens-" + t.Name(),
		WorkerCount: 2,
		DataRoot:    filepath.Join(env.GetBuildDir(), "fs/test", t.Name(), "data"),
		CachesRoot:  filepath.Join(env.GetBuildDir(), "fs/test", t.Name(), "cache"),
		UiPath:      env.GetUIPath(),
	}

	var server *http.Server
	var services = &models.ServicePack{
		Cnf: cnf,
		Log: log.NewLogPackage("", log.DEBUG),
	}

	mondb, err := database.ConnectToMongo(cnf.MongodbUri, cnf.MongodbName)
	require.NoError(t, err)
	err = mondb.Drop(context.Background())
	require.NoError(t, err)

	start := time.Now()
	server = http.NewServer(cnf.RouterHost, cnf.RouterPort, services)
	server.StartupFunc = func() {
		setup.Startup(cnf, services)
	}
	services.StartupChan = make(chan bool)
	go server.Start()

	// Wait for initial startup
	err = waitForStartup(services.StartupChan)
	require.NoError(t, err)

	logger.Debug.Println("Startup took", time.Since(start))
	require.True(t, services.Loaded.Load())

	// Initialize the server as a backup server
	err = services.InstanceService.InitBackup("TEST-BACKUP", coreAddress, coreApiKey)
	logger.ErrTrace(err)
	require.NoError(t, err)

	logger.Debug.Println("Made backup server")

	server.Restart(false)
	logger.Debug.Println("Restarted...")

	// Wait for backup server startup
	err = waitForStartup(services.StartupChan)
	require.NoError(t, err)

	require.Equal(t, models.BackupServerRole, services.InstanceService.GetLocal().Role)

	cores := services.InstanceService.GetCores()
	require.Len(t, cores, 1)

	core := cores[0]

	err = http.WebsocketToCore(core, services)
	logger.ErrTrace(err)

	coreClient := services.ClientService.GetClientByServerId(core.ServerId())
	retries := 0
	for coreClient == nil && retries < 10 {
		retries++
		time.Sleep(time.Millisecond * 100)

		coreClient = services.ClientService.GetClientByServerId(core.ServerId())
	}
	require.NotNil(t, coreClient)
	assert.True(t, coreClient.Active.Load())

	tsk, err := jobs.BackupOne(core, services)
	logger.ErrTrace(err)

	logger.Debug.Println("Started backup task")
	tsk.Wait()
	complete, _ := tsk.Status()
	require.True(t, complete)

	err = tsk.ReadError()
	if err != nil {
		logger.ErrTrace(err)
		t.FailNow()
	}

	services.Server.Stop()
	coreServices.Server.Stop()
}
