package main

// TODO:

// import (
// 	"context"
// 	"math/rand/v2"
// 	"path/filepath"
// 	"testing"
// 	"time"
//
// 	"github.com/ethanrous/weblens/database"
// 	"github.com/ethanrous/weblens/http"
// 	"github.com/ethanrous/weblens/jobs"
// 	"github.com/ethanrous/weblens/models"
// 	"github.com/ethanrous/weblens/modules/config"
// 	"github.com/ethanrous/weblens/service"
// 	"github.com/ethanrous/weblens/modules/errors"
// 	"github.com/rs/zerolog"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )
//
// func waitForStartup(startupChan chan bool) error {
// 	for {
// 		select {
// 		case sig, ok := <-startupChan:
// 			if ok {
// 				startupChan <- sig
// 			} else {
// 				return nil
// 			}
// 		case <-time.After(time.Second * 10):
// 			return errors.Errorf("Startup timeout")
// 		}
// 	}
// }
//
// func TestStartupCore(t *testing.T) {
// 	if testing.Short() {
// 		t.Skipf("skipping %s in short mode", t.Name())
// 	}
//
// 	t.Parallel()
//
// 	// These logs can be very noisy, so we disable them for this test unless debugging
// 	logger := log.NopLogger()
//
// 	var server *http.Server
//
// 	cnf := config.Config{
// 		RouterHost:  config.GetRouterHost(),
// 		RouterPort:  rand.IntN(2000) + 8080,
// 		MongodbUri:  config.GetMongoURI(),
// 		MongodbName: "weblens-" + t.Name(),
// 		WorkerCount: 2,
// 		DataRoot:    filepath.Join(config.GetBuildDir(), "fs/test", t.Name(), "data"),
// 		CachesRoot:  filepath.Join(config.GetBuildDir(), "fs/test", t.Name(), "cache"),
// 		UiPath:      config.GetUIPath(),
// 	}
//
// 	var services = &models.ServicePack{
// 		Cnf: cnf,
// 		Log: logger,
// 	}
//
// 	mondb, err := database.ConnectToMongo(cnf.MongodbUri, cnf.MongodbName, logger)
// 	require.NoError(t, err)
//
// 	err = mondb.Drop(context.Background())
// 	require.NoError(t, err)
//
// 	start := time.Now()
// 	server = http.NewServer(cnf.RouterHost, cnf.RouterPort, services)
// 	services.StartupChan = make(chan bool)
// 	server.StartupFunc = func() {
// 		setup.Startup(cnf, services)
// 	}
// 	go server.Start()
//
// 	if err := waitForStartup(services.StartupChan); err != nil {
// 		t.Fatal(err)
// 	}
//
// 	logger.Debug().Func(func(e *zerolog.Event) { e.Dur("startup_duration", time.Since(start)).Msgf("Init startup complete") })
// 	assert.True(t, services.Loaded.Load())
//
// 	err = service.InitCore(services, "TEST-CORE")
// 	require.NoError(t, err)
//
// 	// Although Restart() is safely synchronous outside of an HTTP request,
// 	// we call it without waiting to allow for our own timeout logic to be used
// 	// services.Server.Restart(false)
// 	// if err := waitForStartup(services.StartupChan); err != nil {
// 	// 	t.Fatal(err)
// 	// }
// 	// logger.Debug().Func(func(e *zerolog.Event) { e.Dur("startup_duration", time.Since(start)).Msgf("Core startup complete") })
// 	// assert.True(t, services.Loaded.Load())
//
// 	owner, err := services.UserService.CreateOwner("test-username", "test-password", "Test Owner")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	err = services.FileService.CreateUserHome(owner)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	// services.Server.Restart(false)
// 	// if err := waitForStartup(services.StartupChan); err != nil {
// 	// 	t.Fatal(err)
// 	// }
// 	// logger.Debug().Func(func(e *zerolog.Event) { e.Dur("startup_duration", time.Since(start)).Msgf("Core restart complete") })
// 	// assert.True(t, services.Loaded.Load())
//
// 	usersTree, err := services.FileService.GetFileTreeByName(service.UsersTreeKey)
// 	require.NoError(t, err)
//
// 	_, err = usersTree.GetRoot().GetChild("test-username")
// 	assert.NoError(t, err)
//
// 	services.Server.Stop()
// }
//
// func TestStartupBackup(t *testing.T) {
// 	if testing.Short() {
// 		t.Skipf("skipping %s in short mode", t.Name())
// 	}
//
// 	t.Parallel()
//
// 	// These logs can be very noisy, so we disable them for this test unless debugging
// 	nop := log.NopLogger()
// 	logger := log.NewZeroLogger()
//
// 	coreServices, err := tests.NewWeblensTestInstance(t.Name(), config.Config{
// 		Role: string(models.CoreServerRole),
// 	}, nop)
// 	require.NoError(t, err)
//
// 	coreKeys, err := coreServices.AccessService.GetKeysByUser(coreServices.UserService.Get("test-username"))
// 	require.NoError(t, err)
// 	coreAPIKey := coreKeys[0].Key
// 	coreAddress := config.GetProxyAddress(coreServices.Cnf)
//
// 	backupConfig := config.ConfigProvider{
// 		RouterHost:  config.GetRouterHost(),
// 		RouterPort:  rand.IntN(2000) + 8080,
// 		MongodbUri:  config.GetMongoURI(),
// 		MongodbName: "weblens-" + t.Name(),
// 		WorkerCount: 2,
// 		DataRoot:    filepath.Join(config.GetBuildDir(), "fs/test", t.Name(), "data"),
// 		CachesRoot:  filepath.Join(config.GetBuildDir(), "fs/test", t.Name(), "cache"),
// 		UiPath:      config.GetUIPath(),
// 	}
//
// 	var server *http.Server
// 	var backupServices = &models.ServicePack{
// 		Cnf: backupConfig,
// 		Log: logger,
// 	}
//
// 	mondb, err := database.ConnectToMongo(backupConfig.MongodbUri, backupConfig.MongodbName, logger)
// 	require.NoError(t, err)
// 	err = mondb.Drop(context.Background())
// 	require.NoError(t, err)
//
// 	start := time.Now()
// 	server = http.NewServer(backupConfig.RouterHost, backupConfig.RouterPort, backupServices)
// 	server.StartupFunc = func() {
// 		setup.Startup(backupConfig, backupServices)
// 	}
// 	backupServices.StartupChan = make(chan bool)
// 	go server.Start()
//
// 	// Wait for initial startup
// 	err = waitForStartup(backupServices.StartupChan)
// 	require.NoError(t, err)
//
// 	logger.Debug().Func(func(e *zerolog.Event) { e.Dur("startup_duration", time.Since(start)).Msgf("Init startup complete") })
// 	require.True(t, backupServices.Loaded.Load())
//
// 	// Initialize the server as a backup server
//
// 	err = service.InitBackup(backupServices, "TEST-BACKUP", coreAddress, coreAPIKey)
// 	require.NoError(t, err)
//
// 	require.Equal(t, models.BackupServerRole, backupServices.InstanceService.GetLocal().Role)
//
// 	cores := backupServices.InstanceService.GetCores()
// 	require.Len(t, cores, 1)
//
// 	core := cores[0]
//
// 	err = http.WebsocketToCore(core, backupServices)
// 	require.NoError(t, err)
//
// 	coreClient := backupServices.ClientService.GetClientByServerID(core.ServerID())
// 	retries := 0
// 	for coreClient == nil && retries < 10 {
// 		retries++
// 		time.Sleep(time.Millisecond * 100)
//
// 		coreClient = backupServices.ClientService.GetClientByServerID(core.ServerID())
// 	}
// 	require.NotNil(t, coreClient)
// 	assert.True(t, coreClient.Active.Load())
//
// 	tsk, err := jobs.BackupOne(core, backupServices)
// 	require.NoError(t, err)
//
// 	logger.Debug().Func(func(e *zerolog.Event) { e.Msgf("Started backup task") })
// 	tsk.Wait()
// 	complete, _ := tsk.Status()
// 	require.True(t, complete)
//
// 	err = tsk.ReadError()
// 	require.NoError(t, err)
//
// 	backupServices.Server.Stop()
// 	coreServices.Server.Stop()
// }
