package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ethanrous/weblens/database"
	"github.com/ethanrous/weblens/http"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/jobs"
	"github.com/ethanrous/weblens/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartupCore(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping %s in short mode", t.Name())
	}

	var server *http.Server
	var services = &models.ServicePack{}

	gin.SetMode(gin.ReleaseMode)

	mondb, err := database.ConnectToMongo(env.GetMongoURI("TEST-CORE"), env.GetMongoDBName("TEST-CORE"))
	require.NoError(t, err)

	err = mondb.Drop(context.Background())
	require.NoError(t, err)

	start := time.Now()
	server = http.NewServer(env.GetRouterHost("TEST-CORE"), env.GetRouterPort("TEST-CORE"), services)
	server.StartupFunc = func() {
		startup("TEST-CORE", services, server)
	}

	services.StartupChan = make(chan bool)
	go server.Start()

	log.Trace.Println("Waiting for core startup...")

StartupWaitLoop:
	for {
		select {
		case _, ok := <-services.StartupChan:
			if ok {
				// If we accidentally catch a real signal, pass it on
				services.StartupChan <- true
			} else {
				// Otherwise, break the loop
				break StartupWaitLoop
			}
		// Allow 10 seconds for the server to start, although it should be much faster
		case <-time.After(time.Second * 10):
			t.Fatal("Core server setup timeout")
		}
	}

	log.Trace.Println("Core startup took", time.Since(start))
	assert.True(t, services.Loaded.Load())

	err = services.InstanceService.InitCore("TEST-CORE")
	log.ErrTrace(err)
	require.NoError(t, err)

	services.Server.Stop()
}

func waitForStartup(startupChan chan bool) error {
	for {
		select {
		case _, ok := <-startupChan:
			if ok {
				startupChan <- true
			} else {
				return nil
			}
		case <-time.After(time.Second * 10):
			return werror.Errorf("Startup timeout")
		}
	}
}

func TestStartupBackup(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping %s in short mode", t.Name())
	}

	if os.Getenv("REMOTE_TESTS") != "true" {
		t.Skipf("skipping %s without REMOTE_TESTS set", t.Name())
	}

	var server *http.Server
	var services = &models.ServicePack{}

	gin.SetMode(gin.ReleaseMode)

	coreAddress := os.Getenv("CORE_ADDRESS")
	require.NotEmpty(t, coreAddress)

	coreApiKey := env.GetCoreApiKey()
	require.NotEmpty(t, coreApiKey)

	mondb, err := database.ConnectToMongo(env.GetMongoURI("TEST-BACKUP"), env.GetMongoDBName("TEST-BACKUP"))
	require.NoError(t, err)
	err = mondb.Drop(context.Background())
	require.NoError(t, err)

	start := time.Now()
	server = http.NewServer(env.GetRouterHost("TEST-BACKUP"), env.GetRouterPort("TEST-BACKUP"), services)
	server.StartupFunc = func() {
		startup("TEST-BACKUP", services, server)
	}
	services.StartupChan = make(chan bool)
	go server.Start()

	log.Debug.Println("Waiting for backup startup...")

	// Wait for initial startup
	err = waitForStartup(services.StartupChan)
	require.NoError(t, err)

	log.Debug.Println("Backup startup complete")

	log.Trace.Println("Startup took", time.Since(start))
	require.True(t, services.Loaded.Load())

	// Initialize the server as a backup server
	err = services.InstanceService.InitBackup("TEST-BACKUP", coreAddress, coreApiKey)
	log.ErrTrace(err)
	require.NoError(t, err)

	server.Restart()

	// Wait for backup server startup
	err = waitForStartup(services.StartupChan)
	require.NoError(t, err)

	require.Equal(t, models.BackupServer, services.InstanceService.GetLocal().Role)

	cores := services.InstanceService.GetCores()
	require.Len(t, cores, 1)

	core := cores[0]

	err = http.WebsocketToCore(core, services)
	log.ErrTrace(err)

	coreClient := services.ClientService.GetClientByServerId(core.ServerId())
	retries := 0
	for coreClient == nil && retries < 5 {
		retries++
		time.Sleep(time.Millisecond * 500)

		coreClient = services.ClientService.GetClientByServerId(core.ServerId())
	}
	require.NotNil(t, coreClient)
	assert.True(t, coreClient.Active.Load())

	tsk, err := jobs.BackupOne(core, services)
	log.ErrTrace(err)

	tsk.Wait()
	complete, _ := tsk.Status()
	require.True(t, complete)

	err = tsk.ReadError()
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	services.Server.Stop()
}
