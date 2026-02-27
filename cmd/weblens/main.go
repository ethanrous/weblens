// Package main is the entry point for the Weblens application server.
package main

import (
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/routers"

	_ "net/http/pprof"
)

func main() {
	// Load configuration. This has a set of defaults, which then can be
	// overridden by environment variables
	cnf := config.GetConfig()

	// This is the main weblens server, we always want to do file discovery
	// when entering through here. Testing is a different story.
	cnf.DoFileDiscovery = true

	// Initialize logger
	logger := log.NewZeroLogger(log.CreateOpts{Level: cnf.LogLevel})

	// Capture interrupt signals to allow for graceful shutdown.
	// The returned context will be canceled on interrupt.
	ctx, cancel := routers.CaptureInterrupt()
	defer cancel()

	err := routers.Start(routers.StartupOpts{
		Ctx:        ctx,
		Cnf:        cnf,
		Logger:     logger,
		CancelFunc: cancel,
	})
	if err != nil {
		logger.Fatal().Stack().Err(err).Msgf("Failed to start Weblens server")
	}
}
