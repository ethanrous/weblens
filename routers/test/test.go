package test

import (
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/rs/zerolog"
)

func NewWeblensTestInstance(testName string, cnf config.ConfigProvider, log zerolog.Logger) (context.AppContext, error) {
	return context.AppContext{}, errors.New("not implemented")
	// var server *http.Server
	//
	// cnf.RouterHost = env.GetRouterHost()
	// cnf.RouterPort = rand.IntN(2000) + 8090
	// cnf.MongodbUri = env.GetMongoURI()
	// cnf.MongodbName = "weblens-" + testName
	// cnf.WorkerCount = 2
	// cnf.DataRoot = filepath.Join(env.GetBuildDir(), "fs/test", testName+"-auto", "data")
	// cnf.CachesRoot = filepath.Join(env.GetBuildDir(), "fs/test", testName+"-auto", "cache")
	// cnf.UiPath = env.GetUIPath()
	//
	// var services = &models.ServicePack{
	// 	Cnf: cnf,
	// 	Log: log,
	// }
	//
	// mondb, err := database.ConnectToMongo(cnf.MongodbUri, cnf.MongodbName, log)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// err = mondb.Drop(context.Background())
	// if err != nil {
	// 	return nil, err
	// }
	//
	// err = mondb.Client().Disconnect(context.Background())
	// if err != nil {
	// 	return nil, err
	// }
	//
	// server = http.NewServer(cnf.RouterHost, cnf.RouterPort, services)
	// server.StartupFunc = func() {
	// 	setup.Startup(cnf, services)
	// }
	// services.StartupChan = make(chan bool)
	// go server.Start()
	//
	// if err := waitForStartup(services.StartupChan); err != nil {
	// 	return nil, err
	// }
	//
	// if models.ServerRole(cnf.Role) == models.CoreServerRole {
	// 	err = services.InstanceService.InitCore(testName)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	owner, err := services.UserService.CreateOwner("test-username", "test-password", "Test Owner")
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	// Although Restart() is safely synchronous outside of an HTTP request,
	// 	// we call it without waiting to allow for our own timeout logic to be used
	// 	services.Server.Restart(false)
	// 	if err := waitForStartup(services.StartupChan); err != nil {
	// 		return nil, err
	// 	}
	//
	// 	_, err = services.AccessService.GenerateApiKey(owner, services.InstanceService.GetLocal(), "test-key")
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// } else if models.ServerRole(cnf.Role) == models.BackupServerRole {
	// 	_, err = services.InstanceService.InitBackup(testName+"-backup", cnf.CoreAddress, cnf.CoreApiKey)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	// Although Restart() is safely synchronous outside of an HTTP request,
	// 	// we call it without waiting to allow for our own timeout logic to be used
	// 	services.Server.Restart(false)
	// 	if err := waitForStartup(services.StartupChan); err != nil {
	// 		return nil, err
	// 	}
	// }
	//
	// return services, nil
}
