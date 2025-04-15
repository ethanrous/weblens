package routers

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/startup"
	v1 "github.com/ethanrous/weblens/routers/api/v1"
	"github.com/ethanrous/weblens/routers/router"
	"github.com/ethanrous/weblens/routers/web"
	context_service "github.com/ethanrous/weblens/services/context"
	file_service "github.com/ethanrous/weblens/services/file"
	"github.com/ethanrous/weblens/services/jobs"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func startupRecover(l zerolog.Logger) {
	if r := recover(); r != nil {
		err := errors.Errorf("%v", r)
		l.Fatal().Stack().Err(err).Msgf("Startup failed:")
	}
}

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

func Startup(ctx context_service.AppContext, cnf config.ConfigProvider) (*router.Router, error) {
	defer startupRecover(ctx.Logger)

	r := router.NewRouter()

	fs.RegisterAbsolutePrefix(file_service.UsersTreeKey, filepath.Join(cnf.DataPath, "users"))
	fs.RegisterAbsolutePrefix(file_service.RestoreTreeKey, filepath.Join(cnf.DataPath, ".restore"))
	fs.RegisterAbsolutePrefix(file_service.CachesTreeKey, cnf.CachePath)

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

	clientService := notify.NewClientManager()
	ctx.ClientService = clientService

	taskService := task.NewWorkerPool(ctx, cnf.WorkerCount)
	jobs.RegisterJobs(taskService)
	taskService.Run()
	ctx.TaskService = taskService

	if err := loadState(ctx); err != nil {
		ctx.Log().Fatal().Stack().Err(err).Msg("Failed to load initial state")
	}

	err = startup.RunStartups(ctx, cnf)
	if err != nil {
		return nil, err
	}

	r.WithAppContext(ctx)

	for _, lm := range router.LoggerMiddlewares(ctx.Logger) {
		r.Use(router.WrapHandlerProvider(lm))
	}

	r.Use(router.Recoverer, router.WeblensAuth)

	r.Mount("/api/v1", v1.Routes)
	r.Mount("/docs", v1.Docs)
	r.Mount("/", web.UiRoutes(web.NewMemFs(ctx, cnf)))

	return r, nil
}

func loadState(ctx context_service.AppContext) error {
	local, err := tower_model.GetLocal(ctx)
	if err != nil {
		ctx.Log().Info().Msgf("No local instance found, creating new one")
		local, err = tower_model.CreateLocal(ctx)
		if err != nil {
			return errors.Wrap(err, "Failed to create local instance")
		}
	}
	ctx.Log().Info().Msgf("Local instance found: %s -- %s", local.TowerId, local.Role)

	if local.Role == tower_model.CoreTowerRole {
		_, err := user_model.GetServerOwner(ctx)
		if err != nil {
			ctx.Log().Warn().Err(err).Msgf("No server owner found, reverting to server init state")
			local, err = tower_model.ResetLocal(ctx)
			if err != nil {
				return errors.Wrap(err, "Failed to reset local instance")
			}
		}

	}

	ctx.LocalTowerId = local.TowerId

	return nil
}

func loadFs(ctx context_service.AppContext) {

}
