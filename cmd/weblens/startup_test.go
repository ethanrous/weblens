package main

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/http"
	"github.com/ethrousseau/weblens/internal/env"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestStartupCore(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping %s in short mode", t.Name())
	}

	prevPort := env.GetRouterPort()
	defer os.Setenv("SERVER_PORT", prevPort)

	port := strconv.Itoa(8090 + (rand.Int() % 1000))
	t.Setenv("SERVER_PORT", port)

	config, err := env.ReadConfig(env.GetConfigName())

	err = setServerState(models.CoreServer)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	gin.SetMode(gin.ReleaseMode)
	start := time.Now()
	server = http.NewServer(config["routerHost"].(string), config["routerPort"].(string), services)
	go server.Start()

	mondb, err := database.ConnectToMongo(env.GetMongoURI(), t.Name())
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}
	err = mondb.Drop(context.Background())
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	startup(config, services, server)
	log.Debug.Println("Startup took", time.Since(start).Seconds())
	assert.True(t, services.Loaded.Load())
}

func TestStartupBackup(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping %s in short mode", t.Name())
	}

	if os.Getenv("REMOTE_TESTS") != "true" {
		t.Skipf("skipping %s without REMOTE_TESTS set", t.Name())
	}

	prevPort := os.Getenv("SERVER_PORT")
	defer os.Setenv("SERVER_PORT", prevPort)

	port := strconv.Itoa(8090 + (rand.Int() % 1000))
	t.Setenv("SERVER_PORT", port)

	err := setServerState(models.BackupServer)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	go main()
	start := time.Now()
	for {
		if services != nil && services.Loaded.Load() &&
			services.ClientService.GetClientByInstanceId("TEST_REMOTE") != nil {
			break
		}

		if time.Since(start) > time.Second*10 {
			t.Fatal("Backup server setup timeout")
		}

		time.Sleep(250 * time.Millisecond)
	}

	coreClient := services.ClientService.GetClientByInstanceId("TEST_REMOTE")
	if !assert.NotNil(t, coreClient) {
		t.FailNow()
	}

	assert.True(t, coreClient.Active.Load())
}

func setServerState(role models.ServerRole) error {
	mongoName := env.GetMongoDBName()
	if !strings.Contains(mongoName, "test") {
		panic(werror.Errorf("MongoDB name (%s) does not include \"test\" during test", mongoName))
	}

	mondb, err := database.ConnectToMongo(env.GetMongoURI(), mongoName)
	if err != nil {
		return werror.WithStack(err)
	}

	servers := mondb.Collection("servers")
	err = servers.Drop(context.Background())
	if err != nil {
		return werror.WithStack(err)
	}

	thisServer := models.NewInstance("TEST_LOCAL", "test", "", role, true, "")
	_, err = servers.InsertOne(context.Background(), thisServer)
	if err != nil {
		return werror.WithStack(err)
	}

	if role == models.BackupServer {
		remoteCore := models.NewInstance(
			"TEST_REMOTE", "test remote", models.WeblensApiKey(
				env.GetCoreApiKey(),
			), models.CoreServer, false, "http://localhost:8089",
		)
		_, err = servers.InsertOne(context.Background(), remoteCore)
		if err != nil {
			return werror.WithStack(err)
		}
	}

	return nil
}
