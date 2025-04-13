package main

import (
	"net/http"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/routers"
	"github.com/ethanrous/weblens/services/context"
)

func main() {
	cnf := config.GetConfig()
	cnf.DoFileDiscovery = true

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
	logger.Fatal().Stack().Err(err).Msg("Router exited unexpectedly")
}
