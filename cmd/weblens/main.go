package main

import (
	"os"

	"github.com/ethanrous/weblens/http"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/setup"
	"github.com/ethanrous/weblens/models"
	"github.com/rs/zerolog"
)

func main() {
	var server *http.Server

	logger := log.NewZeroLogger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	cnf, err := env.GetConfig(env.GetConfigName(), true)
	if err != nil {
		logger.Fatal().Stack().Err(err).Msg("Failed to load config")
		os.Exit(1)
	}

	var services = &models.ServicePack{
		Cnf:         cnf,
		Log:         logger,
		StartupChan: make(chan bool),
	}

	defer setup.MainRecovery(services.Log)
	logger.Info().Msg("Starting Weblens")

	server = http.NewServer(cnf.RouterHost, cnf.RouterPort, services)
	server.StartupFunc = func() {
		setup.Startup(cnf, services)
	}
	server.Start()
}
