// Package config provides configuration management for Weblens server settings.
package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var projectPackagePrefix string
var cnf Provider
var cnfMu sync.RWMutex

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

// Provider provides configuration for Weblens options. All values provided are external to the application, and are expected to be set
// prior to initial startup using environment variables, etc. For management of runtime/mutable server settings, those will be stored in the
// database at /models/settings/...
type Provider struct {
	// Router settings //
	Host         string
	Port         string
	ProxyAddress string

	// Database settings //
	MongoDBUri  string
	MongoDBName string

	// External resource identifiers //
	UIPath            string
	DataPath          string
	CachePath         string
	StaticContentPath string
	// HdirURI is the URI for the HDIR (high dimension image recognition) service, used for machine learning on images to allow semantic search and other features. This is expected to be a separate service, and the endpoint for that service should be provided here.
	HdirURI string

	// Logging settings
	LogLevel  zerolog.Level
	LogFormat string
	LogPath   string

	// Tower auto-initialization config options //
	// If set, the server will attempt to initialize itself with the specified role on first startup.
	InitRole string
	// The address of the Core tower to connect to when initializing as a Backup tower.
	CoreAddress string
	// The API token to use when connecting to the Core tower during initialization as a Backup tower.
	CoreToken string
	// Whether to generate an initial admin API token on first startup (Core server only). Useful for automated testing that uses the API, but should be used with extreme caution (see: never) in production environments.
	GenerateAdminAPIToken bool
	// DangerouslyInsecurePasswordHashing is a flag that, when set to true, will use a much faster bcrypt hashing cost for password hashing. This is intended for testing purposes only, as it significantly reduces the security of stored passwords. It should never be enabled in production environments.
	DangerouslyInsecurePasswordHashing bool

	// Misc config options //
	// BackupInterval specifies how often the server should perform backups of its data. This is only relevant for Backup towers
	BackupInterval time.Duration
	// WorkerCount specifies the number of worker goroutines to use for processing tasks. If set to 0, it will default to the number of CPU cores.
	WorkerCount int
	// DoCache indicates whether to use caching for database operations.
	DoCache bool
	// DoProfile Indicates whether to enable profiling endpoints.
	DoProfile bool
	// DoFileDiscovery Indicates whether to perform file discovery on startup. This is only set to true when running the main Weblens server,
	// and not during tests or other auxiliary binaries.
	DoFileDiscovery bool
}

// Merge merges another Provider into the current one, overriding any non-zero values.
func (c Provider) Merge(o Provider) Provider {
	if o.Host != "" {
		c.Host = o.Host
	}

	if o.Port != "" {
		c.Port = o.Port
	}

	if o.ProxyAddress != "" {
		c.ProxyAddress = o.ProxyAddress
	}

	if o.MongoDBUri != "" {
		c.MongoDBUri = o.MongoDBUri
	}

	if o.MongoDBName != "" {
		c.MongoDBName = o.MongoDBName
	}

	if o.UIPath != "" {
		c.UIPath = o.UIPath
	}

	if o.DataPath != "" {
		c.DataPath = o.DataPath
	}

	if o.CachePath != "" {
		c.CachePath = o.CachePath
	}

	if o.StaticContentPath != "" {
		c.StaticContentPath = o.StaticContentPath
	}

	if o.LogLevel != zerolog.NoLevel {
		c.LogLevel = o.LogLevel
	}

	if o.LogFormat != "" {
		c.LogFormat = o.LogFormat
	}

	if o.WorkerCount != 0 {
		c.WorkerCount = o.WorkerCount
	}

	if o.BackupInterval != 0 {
		c.BackupInterval = o.BackupInterval
	}

	if o.InitRole != "" {
		c.InitRole = o.InitRole
	}

	if o.CoreToken != "" {
		c.CoreToken = o.CoreToken
	}

	if o.CoreAddress != "" {
		c.CoreAddress = o.CoreAddress
	}

	c.DoCache = o.DoCache
	c.DoProfile = o.DoProfile
	c.GenerateAdminAPIToken = o.GenerateAdminAPIToken
	c.DoFileDiscovery = o.DoFileDiscovery

	return c
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

func getDefaultConfig() Provider {
	return Provider{
		Host:              "0.0.0.0",
		Port:              "8080",
		MongoDBUri:        "mongodb://127.0.0.1:27017/?replicaSet=rs0&directConnection=true",
		MongoDBName:       "weblens",
		HdirURI:           "http://weblens-hdir:5000",
		UIPath:            "/app/web",
		StaticContentPath: "/app/static",

		DataPath:  "/data",
		CachePath: "/cache",

		LogLevel:  zerolog.InfoLevel,
		LogFormat: "json",

		WorkerCount:    runtime.NumCPU(),
		BackupInterval: time.Hour,

		DoCache:   true,
		DoProfile: false,
	}
}

func handlePath(path string) string {
	if path[0] == '.' {
		path = filepath.Join(projectPackagePrefix, path)
	}

	return path
}

func getEnvOverride(config *Provider) {
	env := ".env"
	if envPath := os.Getenv("WEBLENS_ENV_PATH"); envPath != "" {
		env = envPath
	}

	err := godotenv.Load(env)
	if err != nil {
		log.Trace().Msgf("No .env file found, using default config: %s", err.Error())
	}

	if logLevel := os.Getenv("WEBLENS_LOG_LEVEL"); logLevel != "" {
		parsedLevel, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			panic(wlerrors.Errorf("invalid log level in WEBLENS_LOG_LEVEL (%s): %w", logLevel, err))
		}

		zerolog.SetGlobalLevel(parsedLevel)

		log.Level(parsedLevel)
		log.Trace().Msgf("Overriding LogLevel with WEBLENS_LOG_LEVEL: %s (parsed: %v)", logLevel, parsedLevel)
		config.LogLevel = parsedLevel

		if config.LogLevel == zerolog.NoLevel {
			config.LogLevel = zerolog.InfoLevel
		}
	}

	if logFormat := os.Getenv("WEBLENS_LOG_FORMAT"); logFormat != "" {
		if logFormat != "dev" {
			logFormat = "json"
		}

		log.Trace().Msgf("Overriding LogFormat with WEBLENS_LOG_FORMAT: %s", logFormat)
		config.LogFormat = logFormat
	}

	if logPath := os.Getenv("WEBLENS_LOG_PATH"); logPath != "" {
		if !filepath.IsAbs(logPath) {
			logPath = filepath.Join(projectPackagePrefix, logPath)
		}

		log.Trace().Msgf("Overriding LogPath with WEBLENS_LOG_PATH: %s", logPath)
		config.LogPath = logPath
	}

	if host := os.Getenv("WEBLENS_HOST"); host != "" {
		log.Trace().Msgf("Overriding Host with WEBLENS_HOST: %s", host)
		config.Host = host
	}

	if port := os.Getenv("WEBLENS_PORT"); port != "" {
		log.Trace().Msgf("Overriding Port with WEBLENS_PORT: %s", port)
		config.Port = port
	}

	if workerCount := os.Getenv("WEBLENS_WORKER_COUNT"); workerCount != "" {
		log.Trace().Msgf("Overriding worker count with WEBLENS_WORKER_COUNT: %s", workerCount)

		if count, err := strconv.Atoi(workerCount); err == nil && count > 0 {
			config.WorkerCount = count
		} else {
			log.Warn().Msgf("Invalid WEBLENS_WORKER_COUNT value: %s, using default worker count: %d", workerCount, config.WorkerCount)
		}
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

		if !filepath.IsAbs(config.DataPath) {
			panic(wlerrors.Errorf("WEBLENS_DATA_PATH must be an absolute path, got: %s", config.DataPath))
		}
	}

	if cachePath := os.Getenv("WEBLENS_CACHE_PATH"); cachePath != "" {
		log.Trace().Msgf("Overriding CachePath with WEBLENS_CACHE_PATH: %s", cachePath)
		config.CachePath = handlePath(cachePath)

		if !filepath.IsAbs(config.CachePath) {
			panic(wlerrors.Errorf("WEBLENS_CACHE_PATH must be an absolute path, got: %s", config.CachePath))
		}
	}

	if doCache, ok := envBool("WEBLENS_DO_CACHE"); ok {
		log.Trace().Msgf("Overriding DoCache with WEBLENS_DO_CACHE: %v", doCache)
		config.DoCache = doCache
	}

	if doProfile, ok := envBool("WEBLENS_DO_PROFILING"); ok {
		log.Trace().Msgf("Overriding DoProfile with WEBLENS_DO_PROFILING: %v", doProfile)
		config.DoProfile = doProfile
	}

	if hdirURI := os.Getenv("WEBLENS_HDIR_URI"); hdirURI != "" {
		log.Trace().Msgf("Overriding HdirURI with WEBLENS_HDIR_URI: %v", hdirURI)
		config.HdirURI = hdirURI
	}

	if doQuickPassHashing, ok := envBool("WEBLENS_USE_DANGEROUSLY_INSECURE_PASSWORD_HASHING"); ok && doQuickPassHashing {
		log.Trace().Msgf("Overriding DangerouslyInsecurePasswordHashing with WEBLENS_USE_DANGEROUSLY_INSECURE_PASSWORD_HASHING: %v", doQuickPassHashing)
		config.DangerouslyInsecurePasswordHashing = doQuickPassHashing
	}
}

// GetConfig returns the current configuration for the Weblens server.
func GetConfig() Provider {
	cnfMu.RLock()
	defer cnfMu.RUnlock()

	return cnf
}

// GetMongoDBUri returns the MongoDB connection URI from the current configuration.
func GetMongoDBUri() string {
	cnfMu.RLock()
	defer cnfMu.RUnlock()

	return cnf.MongoDBUri
}

// SetLogLevel sets the global log level in the configuration.
func SetLogLevel(level zerolog.Level) {
	cnfMu.Lock()
	defer cnfMu.Unlock()

	cnf.LogLevel = level
}
