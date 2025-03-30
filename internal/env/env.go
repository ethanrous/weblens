package env

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Config struct {
	// Required
	MongodbName string `json:"mongodbName"`
	MongodbUri  string `json:"mongodbUri"`
	DataRoot    string `json:"dataRoot"`
	CachesRoot  string `json:"cachesRoot"`
	RouterHost  string `json:"routerHost"`
	LogLevel    string `json:"logLevel"`

	// Testing only
	AppRoot     string `json:"appRoot"`
	UiPath      string `json:"uiPath"`
	Role        string
	CoreAddress string
	CoreApiKey  string
	RouterPort  int `json:"routerPort"`

	// Optional
	WorkerCount int  `json:"workerCount"`
	DetachUi    bool `json:"detachUi"`
}

func GetConfig(configName string, withOverrides bool) (Config, error) {
	configDir := GetConfigPath()
	configFilePath := filepath.Join(configDir, "config.json")
	bs, err := os.ReadFile(configFilePath)
	if err != nil {
		return Config{}, err
	}

	var configs map[string]Config
	err = json.Unmarshal(bs, &configs)
	if err != nil {
		return Config{}, err
	}

	cnf, ok := configs[configName]
	if !ok {
		return Config{}, werror.Errorf("Config %s not found", configName)
	}

	cnf.AppRoot = GetAppRootDir()
	cnf.UiPath = GetUIPath()

	if withOverrides {
		cnf.WorkerCount = GetWorkerCount(cnf)
		cnf.DataRoot = GetDataRoot(cnf)
		cnf.CachesRoot = GetCachesRoot(cnf)
		// cnf.LogLevel = GetLogLevel(string(cnf.LogLevel))
		cnf.MongodbName = GetMongoDBName(cnf)
		cnf.MongodbUri = GetMongoURI()
	}

	return cnf, nil
}

func GetConfigName() string {
	configName := os.Getenv("CONFIG_NAME")
	if configName != "" {
		log.Debug().Msgf("Using config [%s]", configName)
		return configName
	}
	return "PROD"
}

func GetWorkerCount(cnf Config) int {
	poolWorkerCount := os.Getenv("POOL_WORKER_COUNT")
	if poolWorkerCount != "" {
		count, err := strconv.Atoi(poolWorkerCount)
		if err == nil {
			return count
		}
		log.Error().Err(err).Msg("Could not parse POOL_WORKER_COUNT")
	}

	if cnf.WorkerCount > 0 {
		return cnf.WorkerCount
	}

	return runtime.NumCPU() - 2
}

var appRoot string

func GetAppRootDir() string {
	if appRoot != "" {
		return appRoot
	}

	appRoot = os.Getenv("APP_ROOT")
	if appRoot != "" {
		return appRoot
	}

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	appRoot = dir

	if appRoot == "" {
		appRoot = "/app"
	}

	log.Debug().Msgf("AppRoot: %s", appRoot)
	return appRoot
}

func GetUIPath() string {
	// config, err := ReadConfig(GetConfigName())
	// if err != nil {
	// 	panic(err)
	// }
	//
	// uiPath, ok := config["uiPath"].(string)
	// if ok {
	// 	return uiPath
	// }

	// Default
	return filepath.Join(GetAppRootDir(), "ui/dist")
}

func GetRouterPort() string {
	port := os.Getenv("ROUTER_PORT")
	if port == "" {
		return "8080"
	} else {
		return port
	}
}

func GetRouterHost() string {
	host := os.Getenv("ROUTER_HOST")
	if host == "" {
		return "localhost"
	} else {
		return host
	}
}

// func GetLogLevel(level string) log.Level {
// 	envLevel := os.Getenv("LOG_LEVEL")
// 	if envLevel != "" {
// 		level = envLevel
// 	}
//
// 	if level != "" {
// 		switch level {
// 		case "debug":
// 			return log.DEBUG
// 		case "trace":
// 			return log.TRACE
// 		}
// 	}
//
// 	return log.DEFAULT
// }

func GetLogFile() string {
	logPath := os.Getenv("WEBLENS_LOG_FILE")
	if logPath != "" {
		return filepath.Join(GetAppRootDir(), logPath)
	}

	return ""
}

// DetachUi Controls if we host UI files from this router. UI can be hosted elsewhere and
// must proxy any /api/* requests back to this server
func DetachUi() bool {
	detachUi := os.Getenv("DETACH_UI")
	return detachUi == "true"
}

func GetMongoURI() string {
	mongoUri := os.Getenv("MONGODB_URI")
	if mongoUri != "" {
		return mongoUri
	}

	return "mongodb://localhost:27017"
}

func GetMongoDBName(cnf Config) string {
	mongoName := os.Getenv("MONGO_DB_NAME")
	if mongoName != "" {
		return mongoName
	}

	if cnf.MongodbName != "" {
		return cnf.MongodbName
	}

	return "weblens"
}

func GetProxyAddress(cnf Config) string {
	proxyAddress := os.Getenv("PROXY_ADDRESS")
	if proxyAddress != "" {
		return proxyAddress
	}

	proxyAddress = fmt.Sprintf("http://%s:%d", cnf.RouterHost, cnf.RouterPort)
	return proxyAddress
}

func GetTestMediaPath() string {
	testMediaPath := filepath.Join(GetAppRootDir(), "/images/testMedia")
	return testMediaPath
}

func GetConfigPath() string {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath != "" {
		return configPath
	}
	return filepath.Join(GetAppRootDir(), "config")
}

func ReadTypesConfig(target any) error {
	typeJson, err := os.Open(filepath.Join(GetConfigPath(), "mediaType.json"))
	if err != nil {
		return err
	}
	defer func(typeJson *os.File) {
		err = typeJson.Close()
		if err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
	}(typeJson)

	typesBytes, err := io.ReadAll(typeJson)
	err = json.Unmarshal(typesBytes, &target)
	if err != nil {
		return err
	}

	return nil
}

func GetBuildDir() string {
	buildDir := filepath.Join(GetAppRootDir(), "build")
	return buildDir
}

func GetDataRoot(cnf Config) string {
	dataRoot := os.Getenv("DATA_ROOT")
	if dataRoot != "" {
		return dataRoot
	}

	dataRoot = cnf.DataRoot
	if dataRoot == "" {
		// Container default
		return "/data"
	}

	if dataRoot[0] == '.' {
		dataRoot = filepath.Join(GetAppRootDir(), dataRoot)
	}

	return dataRoot
}

func GetCachesRoot(cnf Config) string {
	cachesRoot := os.Getenv("CACHE_ROOT")
	if cachesRoot != "" {
		return cachesRoot
	}

	cachesRoot = cnf.CachesRoot
	if cachesRoot == "" {
		// Container default
		return "/cache"
	}

	if cachesRoot[0] == '.' {
		cachesRoot = filepath.Join(GetAppRootDir(), cachesRoot)
	}

	return cachesRoot
}
