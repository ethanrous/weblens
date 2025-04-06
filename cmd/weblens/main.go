package main

import (
	"net/http"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
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

	logger := log.NewZeroLogger()

	ctx, cancel := routers.CaptureInterrupt()
	defer cancel()

	appCtx := context.NewAppContext(context.NewBasicContext(ctx, logger))

	router, err := routers.Startup(appCtx, cnf)
	if err != nil {
		cancel()
		logger.Fatal().Stack().Err(err).Msg("Failed to start server")
	}

	logger.Info().Msgf("Starting Weblens router at %s:%s", cnf.Host, cnf.Port)
	err = http.ListenAndServe(cnf.Host+":"+cnf.Port, router)
	logger.Fatal().Stack().Err(err).Msg("Failed to start server")

	select {}

}
