// Package routers provides HTTP routing and server startup functionality for the Weblens application.
package routers

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/task"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/startup"
	"github.com/ethanrous/weblens/modules/wlerrors"
	v1 "github.com/ethanrous/weblens/routers/api/v1"
	"github.com/ethanrous/weblens/routers/router"
	"github.com/ethanrous/weblens/routers/web"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	file_service "github.com/ethanrous/weblens/services/file"
	"github.com/ethanrous/weblens/services/jobs"
	"github.com/ethanrous/weblens/services/notify"
	_ "github.com/ethanrous/weblens/services/user" // Required to register user service routes
	"github.com/rs/zerolog"
)

// StartupOpts defines the options for starting the Weblens application server.
type StartupOpts struct {
	Ctx context.Context
	Cnf config.Provider

	Logger     *zerolog.Logger
	CancelFunc context.CancelFunc

	Started chan context_service.AppContext
}

func startupRecover() {
	if r := recover(); r != nil {
		err := wlerrors.Errorf("%v", r)

		log.GlobalLogger().Fatal().Stack().Err(err).Msgf("Startup panicked")
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
			log.GlobalLogger().Info().Msg("Received interrupt signal, shutting down...")
			cancel()
		case <-ctx.Done():
		}

		signal.Reset()
	}()

	return ctx, cancel
}

// Start initializes and starts the Weblens application server.
// The main function calls this to boot up the app, as well as tests.
func Start(opts StartupOpts) error {
	ctx := opts.Ctx
	cnf := opts.Cnf
	logger := opts.Logger
	cancel := opts.CancelFunc

	// Create application context. This will be passed to all services and handlers,
	// and acts as the main dependency injection mechanism.
	appCtx := context_service.NewAppContext(context_service.NewBasicContext(ctx, logger))

	if cnf.DoProfile {
		// Start pprof server for profiling and debugging.
		go func() {
			logger.Debug().Msgf("%v+", http.ListenAndServe("0.0.0.0:6060", nil))
		}()
	}

	// Run startup hooks to initialize services and perform setup tasks.
	// This essentially boots up the entire app.
	appCtx, router, err := startServices(appCtx, cnf)
	if err != nil {
		// If we fail to start up, kill all the services that may have started, and exit.
		logger.Error().Stack().Err(err).Msg("Failed to start server")
		cancel()
		appCtx.WG.Wait()

		return err
	}

	logger.Info().Msgf("Starting Weblens router at %s:%s", cnf.Host, cnf.Port)

	if opts.Started != nil {
		opts.Started <- appCtx
	}

	// Create HTTP server
	server := &http.Server{Addr: cnf.Host + ":" + cnf.Port, Handler: router, ReadTimeout: time.Minute * 5}

	// Ensure graceful shutdown of HTTP server on context cancellation.
	context.AfterFunc(ctx, func() {
		logger.Info().Msg("Shutting down router")

		err := server.Shutdown(context.Background())
		if err != nil {
			logger.Error().Stack().Err(err).Msg("Failed to shutdown router")
		}
	})

	// Start HTTP server. In production, this handles *all* incoming requests. Both for
	// API and web UI.
	err = server.ListenAndServe()
	if err != http.ErrServerClosed {
		cancel()
		logger.Error().Stack().Err(err).Msg("Router exited unexpectedly")
	}

	appCtx.WG.Wait()

	return nil
}

// startServices initializes all application services and configures the HTTP router.
func startServices(appCtx context_service.AppContext, cnf config.Provider) (context_service.AppContext, *router.Router, error) {
	defer startupRecover()

	r := router.NewRouter()

	// Ensure database exists and connect to it
	mongo, err := db.ConnectToMongo(appCtx, cnf.MongoDBUri, cnf.MongoDBName)
	if err != nil {
		return context_service.AppContext{}, nil, wlerrors.Errorf("Failed to connect to MongoDB: %w", err)
	}

	appCtx.DB = mongo

	// Initialize file service
	fileService, err := file_service.NewFileService(appCtx)
	if err != nil {
		return context_service.AppContext{}, nil, wlerrors.Errorf("Failed to initialize file service: %w", err)
	}

	appCtx.FileService = fileService

	// Initialize client notification service (websocket events, etc)
	clientService := notify.NewClientManager(appCtx)
	appCtx.ClientService = clientService

	// Initialize task service
	taskService := task.NewWorkerPool(cnf.WorkerCount)
	jobs.RegisterJobs(taskService)
	taskService.Run(appCtx)
	appCtx.TaskService = taskService

	var local tower_model.Instance

	if local, err = loadLocalTower(appCtx); err != nil {
		return context_service.AppContext{}, nil, wlerrors.Errorf("Failed to load initial state: %w", err)
	}

	if local.TowerID == "" {
		return context_service.AppContext{}, nil, wlerrors.Errorf("Local tower ID is empty after load")
	}

	appCtx.LocalTowerID = local.TowerID

	appCtx = appCtx.WithValue("towerID", local.TowerID)

	if local.Role == tower_model.RoleBackup {
		appCtx = appCtx.WithValue(file_service.SkipJournalKey, true)
	}

	if cnf.InitRole == "" {
		cnf.InitRole = string(local.Role)
	}

	// Run setup functions for various services
	err = startup.RunStartups(appCtx, cnf)
	if err != nil {
		return context_service.AppContext{}, nil, err
	}

	// Install middlewares
	r.Use(
		context_service.AppContexter(appCtx),
		router.CORSMiddleware,
	)

	// Install routes
	r.Mount("/api/v1/", router.LoggerMiddlewares(), router.Recoverer, v1.Routes(appCtx))

	r.Use(router.Recoverer)
	r.Mount("/docs", v1.Docs())
	r.Mount("/", web.UIRoutes(web.NewMemFs(appCtx, cnf)))

	return appCtx, r, nil
}

// loadLocalTower retrieves or initializes the local tower instance and ensures proper server state.
func loadLocalTower(ctx context_service.AppContext) (local tower_model.Instance, err error) {
	local, err = tower_model.GetLocal(ctx)
	if err != nil {
		ctx.Log().Info().Msgf("No local tower found, performing first-time setup")

		local, err = tower_model.CreateLocal(ctx)
		if err != nil {
			return local, wlerrors.Wrap(err, "Failed to create local instance")
		}
	} else {
		ctx.Log().Info().Msgf("Existing local tower found with id [%s] and role [%s]", local.TowerID, local.Role)
	}

	if local.Role != tower_model.RoleUninitialized {
		_, err := user_model.GetServerOwner(ctx)
		if err != nil {
			ctx.Log().Warn().Err(err).Msgf("No server owner found, reverting to server init state")

			local, err = tower_model.ResetLocal(ctx)
			if err != nil {
				return local, wlerrors.Wrap(err, "Failed to reset local instance")
			}
		}
	}

	return local, nil
}
