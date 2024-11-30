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
		startup("TEST-CORE", services)
	}

	services.StartupChan = make(chan bool)
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

	_, err = services.UserService.CreateOwner("test-username", "test-password")
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

	if os.Getenv("REMOTE_TESTS") != "true" {
		t.Skipf("skipping %s without REMOTE_TESTS set", t.Name())
	}
	return

	var server *http.Server
	var services = &models.ServicePack{}

	gin.SetMode(gin.ReleaseMode)

	coreAddress := os.Getenv("CORE_ADDRESS")
	require.NotEmpty(t, coreAddress)

	coreApiKey := os.Getenv("CORE_API_KEY")
	require.NotEmpty(t, coreApiKey)

	mondb, err := database.ConnectToMongo(env.GetMongoURI("TEST-BACKUP"), env.GetMongoDBName("TEST-BACKUP"))
	require.NoError(t, err)
	err = mondb.Drop(context.Background())
	require.NoError(t, err)

	start := time.Now()
	server = http.NewServer(env.GetRouterHost("TEST-BACKUP"), env.GetRouterPort("TEST-BACKUP"), services)
	server.StartupFunc = func() {
		startup("TEST-BACKUP", services)
	}
	services.StartupChan = make(chan bool)
	go server.Start()

	// Wait for initial startup
	err = waitForStartup(services.StartupChan)
	require.NoError(t, err)

	log.Debug.Println("Backup startup complete")

	log.Debug.Println("Startup took", time.Since(start))
	require.True(t, services.Loaded.Load())

	// Initialize the server as a backup server
	err = services.InstanceService.InitBackup("TEST-BACKUP", coreAddress, coreApiKey)
	log.ErrTrace(err)
	require.NoError(t, err)

	server.Restart(false)

	// Wait for backup server startup
	err = waitForStartup(services.StartupChan)
	require.NoError(t, err)

	require.Equal(t, models.BackupServerRole, services.InstanceService.GetLocal().Role)

	cores := services.InstanceService.GetCores()
	require.Len(t, cores, 1)

	core := cores[0]

	err = http.WebsocketToCore(core, services)
	log.ErrTrace(err)

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
