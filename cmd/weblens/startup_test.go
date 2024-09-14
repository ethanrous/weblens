package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/http"
	"github.com/ethrousseau/weblens/internal/env"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/jobs"
	"github.com/ethrousseau/weblens/service/proxy"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartupCore(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping %s in short mode", t.Name())
	}

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
				services.StartupChan <- true
			} else {
				break StartupWaitLoop
			}
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

func TestStartupBackup(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping %s in short mode", t.Name())
	}

	if os.Getenv("REMOTE_TESTS") != "true" {
		t.Skipf("skipping %s without REMOTE_TESTS set", t.Name())
	}

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

	log.Warning.Println("Waiting for backup startup...")

StartupWaitLoop:
	for {
		select {
		case _, ok := <-services.StartupChan:
			if ok {
				services.StartupChan <- true
			} else {
				break StartupWaitLoop
			}
		case <-time.After(time.Second * 10):
			t.Fatal("Backup server setup timeout")
		}
	}

	log.Warning.Println("Backup startup complete")

	log.Trace.Println("Startup took", time.Since(start))
	require.True(t, services.Loaded.Load())

	err = services.InstanceService.InitBackup("TEST-BACKUP", coreAddress, coreApiKey)
	log.ErrTrace(err)
	require.NoError(t, err)

	err = http.WebsocketToCore(services.InstanceService.GetCore(), services)
	log.ErrTrace(err)

	core := services.InstanceService.GetCore()
	services.UserService = proxy.NewProxyUserService(core)

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

	tskErr := tsk.ReadError()
	if err, ok := tskErr.(error); err != nil || (!ok && tskErr != nil) {
		log.ErrTrace(err)
		t.FailNow()
	}

	services.Server.Stop()
}