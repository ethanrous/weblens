// Package routers provides HTTP routing and server startup functionality for the Weblens application.
package routers

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/startup"
	v1 "github.com/ethanrous/weblens/routers/api/v1"
	"github.com/ethanrous/weblens/routers/router"
	"github.com/ethanrous/weblens/routers/web"
	context_service "github.com/ethanrous/weblens/services/context"
	file_service "github.com/ethanrous/weblens/services/file"
	"github.com/ethanrous/weblens/services/jobs"
	"github.com/ethanrous/weblens/services/notify"
	_ "github.com/ethanrous/weblens/services/user" // Required to register user service routes
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func startupRecover(l zerolog.Logger) {
	if r := recover(); r != nil {
		err := errors.Errorf("%v", r)
		l.Fatal().Stack().Err(err).Msgf("Startup failed:")
	}
}

// CaptureInterrupt sets up signal handling for graceful shutdown of the application.
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
			log.Info().Msg("Received interrupt signal, shutting down...")
			cancel()
		case <-ctx.Done():
		}

		signal.Reset()
	}()

	return ctx, cancel
}

// Startup initializes all application services and configures the HTTP router.
func Startup(ctx context_service.AppContext, cnf config.Provider) (*router.Router, error) {
	defer startupRecover(*ctx.Log())

	r := router.NewRouter()

	mongo, err := db.ConnectToMongo(ctx, cnf.MongoDBUri, cnf.MongoDBName)
	if err != nil {
		return nil, err
	}

	ctx.DB = mongo

	fileService, err := file_service.NewFileService(nil)
	if err != nil {
		return nil, err
	}

	ctx.FileService = fileService

	clientService := notify.NewClientManager(ctx)
	ctx.ClientService = clientService

	taskService := task.NewWorkerPool(ctx, cnf.WorkerCount)
	jobs.RegisterJobs(taskService)
	taskService.Run()
	ctx.TaskService = taskService

	var local tower_model.Instance

	if local, err = loadState(ctx); err != nil {
		ctx.Log().Fatal().Stack().Err(err).Msg("Failed to load initial state")
	}

	ctx.LocalTowerID = local.TowerID
	ctx = ctx.WithValue("towerID", local.TowerID)

	if local.Role == tower_model.RoleBackup {
		ctx = ctx.WithValue(file_service.SkipJournalKey, true)
	}

	if cnf.InitRole == "" {
		cnf.InitRole = string(local.Role)
	}

	// Run setup functions for various services
	err = startup.RunStartups(ctx, cnf)
	if err != nil {
		return nil, err
	}

	// Install middlewares
	r.Use(
		context_service.AppContexter(ctx),
		router.CORSMiddleware,
	)

	// Install routes
	r.Mount("/api/v1/", router.LoggerMiddlewares(*ctx.Log()), router.Recoverer, v1.Routes(ctx))

	r.Use(router.Recoverer)
	r.Mount("/docs", v1.Docs())
	r.Mount("/", web.UIRoutes(web.NewMemFs(ctx, cnf)))

	return r, nil
}

func loadState(ctx context_service.AppContext) (local tower_model.Instance, err error) {
	local, err = tower_model.GetLocal(ctx)
	if err != nil {
		ctx.Log().Info().Msgf("No local instance found, creating new one")

		local, err = tower_model.CreateLocal(ctx)
		if err != nil {
			return local, errors.Wrap(err, "Failed to create local instance")
		}
	}

	ctx.Log().Info().Msgf("Local instance found: %s -- %s", local.TowerID, local.Role)

	if local.Role != tower_model.RoleInit {
		_, err := user_model.GetServerOwner(ctx)
		if err != nil {
			ctx.Log().Warn().Err(err).Msgf("No server owner found, reverting to server init state")

			local, err = tower_model.ResetLocal(ctx)
			if err != nil {
				return local, errors.Wrap(err, "Failed to reset local instance")
			}
		}
	}

	return local, nil
}
