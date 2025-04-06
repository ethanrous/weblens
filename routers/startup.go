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
	"github.com/ethanrous/weblens/routers/web"
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

func Startup(ctx *context_service.AppContext, cnf config.ConfigProvider) (*router.Router, error) {
	r := router.NewRouter()

	mongo, err := db.ConnectToMongo(ctx, cnf.MongoDBUri, cnf.MongoDBName)
	if err != nil {
		return nil, err
	}

	ctx.DB = mongo

	r.WithAppContext(*ctx)

	r.Mount("/api/v1/", v1.Routes)
	r.Mount("/docs", v1.Docs)
	r.Mount("/", web.UiRoutes(web.NewMemFs(ctx, cnf)))

	return r, nil
}
