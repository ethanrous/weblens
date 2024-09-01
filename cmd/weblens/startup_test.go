package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func init() {
	internal.SetAppRoot("/Users/ethan/repos/weblens/")

	err := os.Unsetenv("MONGODB_NAME")
	if err != nil {
		panic(err)
	}
	err = os.Unsetenv("SERVER_PORT")
	if err != nil {
		panic(err)
	}

	err = godotenv.Load(filepath.Join(internal.GetConfigDir(), "core-test.env"))
	if err != nil {
		log.Warning.Println("Could not load core-test.env file", err)
	}
}

func TestStartupCore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping core startup test in short mode")
	}

	t.Setenv("SERVER_PORT", "8084")

	err := setServerState(models.CoreServer)
	if err != nil {
		log.ErrTrace(err)
		t.FailNow()
	}

	go main()
	start := time.Now()
	for time.Since(start) < time.Second*5 {
		if services != nil && services.Loaded.Load() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !assert.Less(t, time.Since(start), 5*time.Second) {
		t.FailNow()
	}
}

func TestStartupBackup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping backup startup test in short mode")
	}

	t.Setenv("SERVER_PORT", "8085")

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

	mondb, err := database.ConnectToMongo(internal.GetMongoURI(), internal.GetMongoDBName())
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
			), models.CoreServer, false, "http://localhost:8080",
		)
		_, err = servers.InsertOne(context.Background(), remoteCore)
		if err != nil {
			return werror.WithStack(err)
		}
	}

	return nil
}
