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

	HdirUri string

	UIPath            string
	DataPath          string
	CachePath         string
	StaticContentPath string

	LogLevel        zerolog.Level
	LogFormat       string
	BackupInterval  time.Duration
	WorkerCount     int
	DoCache         bool
	InitRole        string
	DoFileDiscovery bool
}

func envBool(key string) (val bool, ok bool) {
	if value, exists := os.LookupEnv(key); exists {
		if value == "true" {
			return true, true
		}

		return false, true
	}

	return false, false
}

func getDefaultConfig() ConfigProvider {
	return ConfigProvider{
		Host:              "0.0.0.0",
		Port:              "8080",
		MongoDBUri:        "mongodb://weblens-mongo:27017/?replicaSet=rs0",
		MongoDBName:       "weblensDB",
		HdirUri:           "http://weblens-hdir:5000",
		UIPath:            "/app/web",
		StaticContentPath: "/app/static",

		DataPath:  "/data",
		CachePath: "/cache",

		LogLevel:  zerolog.InfoLevel,
		LogFormat: "json",

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
		log.Trace().Msgf("No .env file found, using default config: %s", err.Error())
	}

	if host := os.Getenv("WEBLENS_HOST"); host != "" {
		log.Trace().Msgf("Overriding Host with WEBLENS_HOST: %s", host)
		config.Host = host
	}

	if port := os.Getenv("WEBLENS_PORT"); port != "" {
		log.Trace().Msgf("Overriding Port with WEBLENS_PORT: %s", port)
		config.Port = port
	}

	if proxyAddress := os.Getenv("WEBLENS_PROXY_ADDRESS"); proxyAddress != "" {
		log.Trace().Msgf("Overriding ProxyAddress with WEBLENS_PROXY_ADDRESS: %s", proxyAddress)
		config.ProxyAddress = proxyAddress
	}

	if uiPath := os.Getenv("WEBLENS_UI_PATH"); uiPath != "" {
		log.Trace().Msgf("Overriding UIPath with WEBLENS_UI_PATH: %s", uiPath)
		config.UIPath = handlePath(uiPath)
	}

	if staticContentPath := os.Getenv("WEBLENS_STATIC_CONTENT_PATH"); staticContentPath != "" {
		log.Trace().Msgf("Overriding StaticContentPath with WEBLENS_STATIC_CONTENT_PATH: %s", staticContentPath)
		config.StaticContentPath = staticContentPath
	}

	if mongoDBUri := os.Getenv("WEBLENS_MONGODB_URI"); mongoDBUri != "" {
		log.Trace().Msgf("Overriding MongoDBUri with WEBLENS_MONGODB_URI: %s", mongoDBUri)
		config.MongoDBUri = mongoDBUri
	}

	if mongoDBName := os.Getenv("WEBLENS_MONGODB_NAME"); mongoDBName != "" {
		log.Trace().Msgf("Overriding MongoDBName with WEBLENS_MONGODB_NAME: %s", mongoDBName)
		config.MongoDBName = mongoDBName
	}

	if initRole := os.Getenv("WEBLENS_INIT_ROLE"); initRole != "" {
		log.Trace().Msgf("Overriding InitRole with WEBLENS_INIT_ROLE: %s", initRole)
		config.InitRole = initRole
	}

	if dataPath := os.Getenv("WEBLENS_DATA_PATH"); dataPath != "" {
		log.Trace().Msgf("Overriding DataPath with WEBLENS_DATA_PATH: %s", dataPath)
		config.DataPath = handlePath(dataPath)
	}

	if cachePath := os.Getenv("WEBLENS_CACHE_PATH"); cachePath != "" {
		log.Trace().Msgf("Overriding CachePath with WEBLENS_CACHE_PATH: %s", cachePath)
		config.CachePath = handlePath(cachePath)
	}

	if doCache, ok := envBool("WEBLENS_DO_CACHE"); ok {
		log.Trace().Msgf("Overriding DoCache with WEBLENS_DO_CACHE: %v", doCache)
		config.DoCache = doCache
	}

	if logFormat := os.Getenv("WEBLENS_LOG_FORMAT"); logFormat != "" {
		if logFormat != "dev" {
			logFormat = "json"
		}

		log.Trace().Msgf("Overriding LogFormat with WEBLENS_LOG_FORMAT: %s", logFormat)
		config.LogFormat = logFormat
	}

	if logLevel := os.Getenv("WEBLENS_LOG_LEVEL"); logLevel != "" {
		parsedLevel, _ := zerolog.ParseLevel(logLevel)
		log.Trace().Msgf("Overriding LogLevel with WEBLENS_LOG_LEVEL: %s (parsed: %v)", logLevel, parsedLevel)
		config.LogLevel = parsedLevel

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
