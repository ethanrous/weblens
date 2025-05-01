package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var projectPackagePrefix string
var cnf ConfigProvider

func init() {
	_, filename, _, _ := runtime.Caller(0)

	projectPackagePrefix = strings.TrimSuffix(filename, "modules/config/config.go")
	if projectPackagePrefix == filename {
		// in case the source code file is moved, we can not trim the suffix, the code above should also be updated.
		panic("weblens config unable to detect correct package prefix, please update file: " + filename)
	}

	cnf = getDefaultConfig()
	getEnvOverride(&cnf)
}

// ConfigProvider provides configuration for Weblens options. All values provided are external to the application, and are expected to be set
// prior to initial startup using environment variables, etc. For management of runtime/mutable server settings, those will be stored in the
// database at /models/settings/...
type ConfigProvider struct {
	Host         string
	Port         string
	ProxyAddress string

	MongoDBUri  string
	MongoDBName string

	UIPath            string
	DataPath          string
	CachePath         string
	StaticContentPath string

	LogLevel        zerolog.Level
	BackupInterval  time.Duration
	WorkerCount     int
	DoCache         bool
	InitRole        string
	DoFileDiscovery bool
}

func getDefaultConfig() ConfigProvider {
	return ConfigProvider{
		Host:              "0.0.0.0",
		Port:              "8080",
		MongoDBUri:        "mongodb://localhost:27017/?replicaSet=rs0",
		MongoDBName:       "weblensDB",
		UIPath:            "/app/web",
		StaticContentPath: "/app/web/static",

		DataPath:  "/data",
		CachePath: "/cache",

		LogLevel: zerolog.InfoLevel,

		WorkerCount:    runtime.NumCPU(),
		BackupInterval: time.Hour,

		DoCache: true,
	}
}

func handlePath(path string) string {
	if path[0] == '.' {
		path = filepath.Join(projectPackagePrefix, path)
	}

	return path
}

func getEnvOverride(config *ConfigProvider) {
	env := ".env"
	if envPath := os.Getenv("WEBLENS_ENV_PATH"); envPath != "" {
		env = envPath
	}

	err := godotenv.Load(env)
	if err != nil {
		log.Debug().Msgf("No .env file found, using default config: %s", err.Error())
	}

	if host := os.Getenv("WEBLENS_HOST"); host != "" {
		config.Host = host
	}

	if port := os.Getenv("WEBLENS_PORT"); port != "" {
		config.Port = port
	}

	if proxyAddress := os.Getenv("WEBLENS_PROXY_ADDRESS"); proxyAddress != "" {
		config.ProxyAddress = proxyAddress
	}

	if uiPath := os.Getenv("WEBLENS_UI_PATH"); uiPath != "" {
		config.UIPath = handlePath(uiPath)
	}

	if staticContentPath := os.Getenv("WEBLENS_STATIC_CONTENT_PATH"); staticContentPath != "" {
		config.StaticContentPath = staticContentPath
	}

	if mongoDBUri := os.Getenv("WEBLENS_MONGODB_URI"); mongoDBUri != "" {
		config.MongoDBUri = mongoDBUri
	}

	if mongoDBName := os.Getenv("WEBLENS_MONGODB_NAME"); mongoDBName != "" {
		config.MongoDBName = mongoDBName
	}

	if initRole := os.Getenv("WEBLENS_INIT_ROLE"); initRole != "" {
		config.InitRole = initRole
	}

	if dataPath := os.Getenv("WEBLENS_DATA_PATH"); dataPath != "" {
		config.DataPath = handlePath(dataPath)
	}

	if cachePath := os.Getenv("WEBLENS_CACHE_PATH"); cachePath != "" {
		config.CachePath = handlePath(cachePath)
	}

	if logLevel := os.Getenv("WEBLENS_LOG_LEVEL"); logLevel != "" {
		config.LogLevel, _ = zerolog.ParseLevel(logLevel)
		if config.LogLevel == zerolog.NoLevel {
			config.LogLevel = zerolog.InfoLevel
		}
	}
}

func GetConfig() ConfigProvider {
	return cnf
}

func GetMongoDBUri() string {
	return cnf.MongoDBUri
}
