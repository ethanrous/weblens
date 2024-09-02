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
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/stretchr/testify/assert"
)

func init() {
	internal.ReadEnv()
}

func TestStartupCore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping core startup test in short mode")
	}

	prevPort := internal.GetRouterPort()
	defer os.Setenv("SERVER_PORT", prevPort)

	port := strconv.Itoa(8090 + (rand.Int() % 1000))
	t.Setenv("SERVER_PORT", port)

	err := setServerState(models.CoreServer)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	start := time.Now()
	server = http.NewServer(services)
	go server.Start()

	mondb, err := database.ConnectToMongo(internal.GetMongoURI(), t.Name())
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}
	err = mondb.Drop(context.Background())
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	startup(t.Name(), services, server)
	log.Debug.Println("Startup took", time.Since(start).Seconds())
	assert.True(t, services.Loaded.Load())
}

func TestStartupBackup(t *testing.T) {
	if os.Getenv("REMOTE_TESTS") != "true" {
		t.Skip("REMOTE_TESTS not set")
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
	mongoName := internal.GetMongoDBName()
	if !strings.Contains(mongoName, "test") {
		panic(werror.Errorf("MongoDB name (%s) does not include \"test\" during test", mongoName))
	}

	mondb, err := database.ConnectToMongo(internal.GetMongoURI(), mongoName)
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
				internal.GetCoreApiKey(),
			), models.CoreServer, false, "http://localhost:8089",
		)
		_, err = servers.InsertOne(context.Background(), remoteCore)
		if err != nil {
			return werror.WithStack(err)
		}
	}

	return nil
}
