package tests

import (
	"context"
	"math/rand/v2"
	"path/filepath"
	"time"

	"github.com/ethanrous/weblens/database"
	"github.com/ethanrous/weblens/http"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/setup"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

// Create a new instance of the weblens application to test against
func NewWeblensTestInstance(testName string, cnf env.Config, log *zerolog.Logger) (*models.ServicePack, error) {
	var server *http.Server

	cnf.RouterHost = env.GetRouterHost()
	cnf.RouterPort = rand.IntN(2000) + 8090
	cnf.MongodbUri = env.GetMongoURI()
	cnf.MongodbName = "weblens-" + testName
	cnf.WorkerCount = 2
	cnf.DataRoot = filepath.Join(env.GetBuildDir(), "fs/test", testName+"-auto", "data")
	cnf.CachesRoot = filepath.Join(env.GetBuildDir(), "fs/test", testName+"-auto", "cache")
	cnf.UiPath = env.GetUIPath()

	var services = &models.ServicePack{
		Cnf: cnf,
		Log: log,
	}

	mondb, err := database.ConnectToMongo(cnf.MongodbUri, cnf.MongodbName, log)
	if err != nil {
		return nil, err
	}

	err = mondb.Drop(context.Background())
	if err != nil {
		return nil, err
	}

	err = mondb.Client().Disconnect(context.Background())
	if err != nil {
		return nil, err
	}

	server = http.NewServer(cnf.RouterHost, cnf.RouterPort, services)
	server.StartupFunc = func() {
		setup.Startup(cnf, services)
	}
	services.StartupChan = make(chan bool)
	go server.Start()

	if err := waitForStartup(services.StartupChan); err != nil {
		return nil, err
	}

	if models.ServerRole(cnf.Role) == models.CoreServerRole {
		err = services.InstanceService.InitCore(testName)
		if err != nil {
			return nil, err
		}

		owner, err := services.UserService.CreateOwner("test-username", "test-password", "Test Owner")
		if err != nil {
			return nil, err
		}

		// Although Restart() is safely synchronous outside of an HTTP request,
		// we call it without waiting to allow for our own timeout logic to be used
		services.Server.Restart(false)
		if err := waitForStartup(services.StartupChan); err != nil {
			return nil, err
		}

		_, err = services.AccessService.GenerateApiKey(owner, services.InstanceService.GetLocal(), "test-key")
		if err != nil {
			return nil, err
		}
	} else if models.ServerRole(cnf.Role) == models.BackupServerRole {
		_, err = services.InstanceService.InitBackup(testName+"-backup", cnf.CoreAddress, cnf.CoreApiKey)
		if err != nil {
			return nil, err
		}

		// Although Restart() is safely synchronous outside of an HTTP request,
		// we call it without waiting to allow for our own timeout logic to be used
		services.Server.Restart(false)
		if err := waitForStartup(services.StartupChan); err != nil {
			return nil, err
		}
	}

	return services, nil
}

func waitForStartup(startupChan chan bool) error {
	zlog.Debug().Func(func(e *zerolog.Event) { e.Msgf("Waiting for startup...") })
	for {
		select {
		case sig, ok := <-startupChan:
			if ok {
				zlog.Debug().Func(func(e *zerolog.Event) { e.Msgf("Relaying startup signal") })
				startupChan <- sig
			} else {
				return nil
			}
		case <-time.After(time.Second * 10):
			return werror.Errorf("Startup timeout")
		}
	}
}

func CheckDropCol(col *mongo.Collection, logger *zerolog.Logger) {
	if err := col.Drop(context.Background()); err != nil {
		err = errors.Wrap(err, "Failed to drop collection")
		logger.Error().Stack().Err(err).Msg("Failed to drop collection")
	}
}
