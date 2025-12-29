// Package main is the entry point for the Weblens application server.
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/routers"
	context_service "github.com/ethanrous/weblens/services/context"

	_ "net/http/pprof"
)

func main() {
	cnf := config.GetConfig()
	cnf.DoFileDiscovery = true

	logger := log.NewZeroLogger()

	ctx, cancel := routers.CaptureInterrupt()
	defer cancel()

	appCtx := context_service.NewAppContext(context_service.NewBasicContext(ctx, logger))

	go func() {
		logger.Debug().Msgf("%v+", http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	router, err := routers.Startup(appCtx, cnf)
	if err != nil {
		cancel()
		logger.Fatal().Stack().Err(err).Msg("Failed to start server")
	}

	logger.Info().Msgf("Starting Weblens router at %s:%s", cnf.Host, cnf.Port)

	server := &http.Server{Addr: cnf.Host + ":" + cnf.Port, Handler: router, ReadTimeout: time.Minute * 5}

	context.AfterFunc(ctx, func() {
		logger.Info().Msg("Shutting down router")

		err := server.Shutdown(context.Background())
		if err != nil {
			logger.Error().Stack().Err(err).Msg("Failed to shutdown router")
		}
	})

	err = server.ListenAndServe()
	if err != http.ErrServerClosed {
		cancel()
		logger.Error().Stack().Err(err).Msg("Router exited unexpectedly")
	}

	appCtx.WG.Wait()
}