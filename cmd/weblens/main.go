package main

import (
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/routers"
	"github.com/ethanrous/weblens/services/context"
)

func main() {
	// var server *http.Server
	//
	// logger := log.NewZeroLogger()
	// zerolog.SetGlobalLevel(zerolog.DebugLevel)
	//
	// cnf, err := env.GetConfig(env.GetConfigName(), true)
	// if err != nil {
	// 	logger.Fatal().Stack().Err(err).Msg("Failed to load config")
	// 	os.Exit(1)
	// }
	//
	// var services = &models.ServicePack{
	// 	Cnf:         cnf,
	// 	Log:         logger,
	// 	StartupChan: make(chan bool),
	// }
	//
	// defer setup.MainRecovery(services.Log)
	// logger.Info().Msg("Starting Weblens")
	//
	// server = http.NewServer(cnf.RouterHost, cnf.RouterPort, services)
	// server.StartupFunc = func() {
	// 	setup.Startup(cnf, services)
	// }
	// server.Start()

	cnf := config.GetConfig()
	ctx, cancel := routers.CaptureInterrupt()
	defer cancel()

	routers.Startup(context.AppContext{Context: ctx}, cnf)

	select {
	case <-ctx.Done():
	case <-make(chan struct{}):

		// default:
		// 	fmt.Println("Hello, World!")
	}

}
