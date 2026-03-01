// Package e2e_test provides common utilities and configurations for end-to-end tests.
package e2e_test

import (
	"context"
	"encoding/base64"

	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/bcrypt"

	openapi "github.com/ethanrous/weblens/api"
	"github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/cryptography"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlog"
	"github.com/ethanrous/weblens/routers"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/rs/zerolog"
)

var logLevel = zerolog.DebugLevel

var portsInUse = make(map[int]struct{})
var atomicPort = atomic.Int32{}

type setupResult struct {
	cnf     config.Provider
	address string
	ctx     context_service.AppContext

	// Only set if the server was initialized as a core tower, with an admin token.
	token string
}

func safeTestName(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	strings.ReplaceAll(name, ".", "-")

	return name
}

func buildTestConfig(testName string, override ...config.Provider) config.Provider {
	cnf := config.GetConfig()
	for _, o := range override {
		cnf = cnf.Merge(o)
	}

	testName = safeTestName(testName)
	if cnf.InitRole != "" {
		testName = testName + "-" + cnf.InitRole
	}

	cnf.MongoDBName = testName

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	repoRoot := filepath.Dir(cwd)
	testRoot := repoRoot + "/_build/fs/test/" + testName
	cnf.DataPath = testRoot + "/data"
	cnf.CachePath = testRoot + "/cache"
	cnf.UIPath = repoRoot + "/weblens-vue/weblens-nuxt/.output/public/"

	logPath := "../_build/logs/e2e-test-backends/" + safeTestName(testName) + ".log"
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		panic(err)
	}

	cnf.LogPath = logPath

	err = os.RemoveAll(testRoot)
	if err != nil {
		panic(err)
	}

	port := atomicPort.Add(1) + 14000
	cnf.Port = strconv.Itoa(int(port))

	cnf.WorkerCount = 2

	return cnf
}

func setupTestServer(ctx context.Context, name string, settings ...config.Provider) (setupResult, error) {
	cnf := buildTestConfig(name, settings...)

	ctx, cancel := context.WithCancel(ctx)
	startedChan := make(chan context_service.AppContext)
	failedChan := make(chan error, 1)

	ctx = context.WithValue(ctx, cryptography.BcryptDifficultyCtxKey, bcrypt.MinCost)

	logger := wlog.NewZeroLogger(wlog.CreateOpts{Level: logLevel, LogFile: cnf.LogPath}).With().Str("test", name).Logger()
	appCtx := context_service.NewAppContext(context_service.NewBasicContext(ctx, &logger))

	testDB, err := db.ConnectToMongo(appCtx, cnf.MongoDBUri, cnf.MongoDBName)
	if err != nil {
		cancel()

		return setupResult{}, wlerrors.Wrapf(err, "failed to connect to test database [%s]", cnf.MongoDBName)
	}

	err = testDB.Drop(appCtx)
	if err != nil {
		cancel()

		return setupResult{}, wlerrors.Wrapf(err, "failed to drop mongo test db: [%s]", cnf.MongoDBName)
	}

	context.AfterFunc(ctx, func() {
		logger.Info().Msgf("Cleaning up test database [%s]", cnf.MongoDBName)

		appCtx := context_service.NewAppContext(context_service.NewBasicContext(context.Background(), &logger))

		testDB, err := db.ConnectToMongo(appCtx, cnf.MongoDBUri, cnf.MongoDBName)
		if err != nil {
			logger.Error().Stack().Err(err).Msgf("failed to connect to mongo test db during cleanup: [%s]", cnf.MongoDBName)

			return
		}

		err = testDB.Drop(appCtx)
		if err != nil {
			logger.Error().Stack().Err(err).Msgf("failed to drop mongo test db during cleanup: [%s]", cnf.MongoDBName)
		}
	})

	go func() {
		err := routers.Start(routers.StartupOpts{
			Ctx:        ctx,
			Cnf:        cnf,
			Logger:     &logger,
			CancelFunc: cancel,
			Started:    startedChan,
		})
		failedChan <- err
	}()

	select {
	case err := <-failedChan:
		return setupResult{}, wlerrors.Errorf("%s test server failed to start (logs: %s): %w", cnf.InitRole, cnf.LogPath, err)
	case appCtx = <-startedChan:
	}

	ret := setupResult{
		cnf:     cnf,
		address: "http://localhost:" + cnf.Port,
		ctx:     appCtx,
	}

	if cnf.InitRole == string(tower.RoleCore) && cnf.GenerateAdminAPIToken {
		tokens, err := auth.GetTokensByUser(appCtx, "admin")
		if err != nil {
			return setupResult{}, wlerrors.Errorf("failed to get admin tokens: %w", err)
		}

		if len(tokens) == 0 {
			return setupResult{}, wlerrors.New("no admin tokens found after initialization")
		}

		ret.token = base64.StdEncoding.EncodeToString(tokens[0].Token[:])

		client := getAPIClientFromConfig(cnf, ret.token)

		retries := 0
		for retries < 10 {
			_, resp, err := client.UsersAPI.GetUser(appCtx).Execute()
			if err == nil {
				break
			}

			// Server is started, but gave an error response. This means the server is ready,
			// but something is wrong with the request (likely missing auth token). Don't keep retrying in this case.
			if resp != nil && resp.StatusCode < 500 && resp.StatusCode >= 400 {
				return setupResult{}, wlerrors.Errorf("API returned error response (logs: %s): %w", cnf.LogPath, err)
			}

			logger.Warn().Err(err).Msgf("API not ready yet, retrying... (%d/10)", retries+1)
			retries++

			time.Sleep(1 * time.Second)
		}

		if retries == 10 {
			return setupResult{}, wlerrors.Errorf("API did not become ready in time (logs: %s)", cnf.LogPath)
		}
	}

	logger.Info().Msgf("Test server for test [%s] started on port %s", name, cnf.Port)

	return ret, nil
}

func getAPIClientFromConfig(cnf config.Provider, token string) *openapi.APIClient {
	apiConfig := openapi.NewConfiguration()
	apiConfig.Host = "localhost:" + cnf.Port
	apiConfig.DefaultHeader = make(map[string]string)
	apiConfig.UserAgent = "Weblens-Client-Go"

	if token != "" {
		apiConfig.AddDefaultHeader("Authorization", "Bearer "+string(token))
	}

	client := openapi.NewAPIClient(apiConfig)

	return client
}
