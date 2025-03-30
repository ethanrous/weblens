package routers

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/config"
	v1 "github.com/ethanrous/weblens/routers/api/v1"
	"github.com/ethanrous/weblens/routers/router"
	context_service "github.com/ethanrous/weblens/services/context"
)

func CaptureInterrupt() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// install notify
		signalChannel := make(chan os.Signal, 1)

		signal.Notify(
			signalChannel,
			// ctl-c
			syscall.SIGINT,

			// Docker stop
			syscall.SIGTERM,
		)
		select {
		case <-signalChannel:
		case <-ctx.Done():
		}
		cancel()
		signal.Reset()
	}()

	return ctx, cancel
}

func Startup(ctx context_service.BasicContext, cnf config.ConfigProvider) error {
	r := router.NewRouter()

	mongo, err := db.ConnectToMongo(ctx, cnf.MongoDBUri, cnf.MongoDBName)
	if err != nil {
		return err
	}

	r.Inject(router.Injection{
		DB:  mongo,
		Log: ctx.Logger,
	})

	r.Mount("/api/v1/", v1.Routes())

	return nil
}
