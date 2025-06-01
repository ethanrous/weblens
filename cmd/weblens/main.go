package main

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"runtime/trace"
	"time"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/routers"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/rs/zerolog"
)

func main() {
	file, _ := os.Create("/data/trace.out")
	_ = trace.Start(file)
	defer trace.Stop()

	cnf := config.GetConfig()
	cnf.DoFileDiscovery = true

	logger := log.NewZeroLogger()

	ctx, cancel := routers.CaptureInterrupt()
	defer cancel()

	appCtx := context_service.NewAppContext(context_service.NewBasicContext(ctx, logger))

	router, err := routers.Startup(appCtx, cnf)
	if err != nil {
		cancel()
		logger.Fatal().Stack().Err(err).Msg("Failed to start server")
	}

	logger.Info().Msgf("Starting Weblens router at %s:%s", cnf.Host, cnf.Port)

	server := &http.Server{Addr: cnf.Host + ":" + cnf.Port, Handler: router}

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

	lastCount := -1

	plateauCount := 0
	for plateauCount < 5 {
		newCount := runtime.NumGoroutine()
		if newCount == lastCount {
			plateauCount++
		} else {
			logger.Debug().Msgf("Waiting for %d goroutines to finish", newCount)

			plateauCount = 0
		}

		lastCount = newCount

		// Yield the processor to allow other goroutines to run
		time.Sleep(100 * time.Millisecond)
		runtime.Gosched()
	}

	appCtx.Log().Trace().Func(func(e *zerolog.Event) {
		remaining := runtime.NumGoroutine()
		buf := make([]byte, remaining*1024)
		runtime.Stack(buf, true)
		e.Msgf("Stack trace: %s", buf)
	})
}
