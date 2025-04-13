package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var projectPackagePrefix string

func init() {

	_, filename, _, _ := runtime.Caller(0)
	projectPackagePrefix = strings.TrimSuffix(filename, "modules/config/config.go")
	if projectPackagePrefix == filename {
		// in case the source code file is moved, we can not trim the suffix, the code above should also be updated.
		panic("weblens config unable to detect correct package prefix, please update file: " + filename)
	}
}

// ConfigProvider provides configuration for Weblens options. All values provided are external to the application, and are expected to be set
// prior to initial startup using environment variables, etc. For management of runtime/mutable server settings, those will be stored in the
// database at /models/settings/...
type ConfigProvider struct {
	Host string
	Port string

	ProxyAddress string

	UIPath            string
	StaticContentPath string

	MongoDBUri  string
	MongoDBName string

	InitRole string

	DataPath  string
	CachePath string

	DoFileDiscovery bool

	WorkerCount int
}

func getDefaultConfig() ConfigProvider {
	return ConfigProvider{
		Host:              "localhost",
		Port:              "8080",
		MongoDBUri:        "mongodb://localhost:27017/?replicaSet=rs0",
		MongoDBName:       "weblensDB",
		UIPath:            "/web",
		StaticContentPath: "/web/static",

		DataPath:  "/data",
		CachePath: "/cahhe",

		WorkerCount: runtime.NumCPU(),
	}
}

func handlePath(path string) string {
	if path[0] == '.' {
		path = filepath.Join(projectPackagePrefix, path)
	}
	return path
}

func getEnvOverride(config *ConfigProvider) {
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
}

func GetConfig() ConfigProvider {
	cnf := getDefaultConfig()
	getEnvOverride(&cnf)
	return cnf
}

func GetMongoDBUri() string {
	cnf := GetConfig()
	return cnf.MongoDBUri
}
