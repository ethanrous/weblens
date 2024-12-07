package main

import (
	_ "net/http/pprof"
	"os"

	"github.com/ethanrous/weblens/http"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/setup"
	"github.com/ethanrous/weblens/models"
)

func main() {
	var server *http.Server

	cnf, err := env.GetConfig(env.GetConfigName(), true)
	if err != nil {
		log.ErrTrace(err)
		os.Exit(1)
	}

	var services = &models.ServicePack{
		Cnf: cnf,
		Log: log.NewLogPackage("", log.Level(cnf.LogLevel)),
	}

	defer setup.MainRecovery("WEBLENS ENCOUNTERED AN UNRECOVERABLE ERROR")
	log.Info.Println("Starting Weblens")

	server = http.NewServer(cnf.RouterHost, cnf.RouterPort, services)
	server.StartupFunc = func() {
		setup.Startup(cnf, services)
	}
	services.StartupChan = make(chan bool)
	server.Start()
}
