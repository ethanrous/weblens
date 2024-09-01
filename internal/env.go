package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/ethrousseau/weblens/internal/log"
)

func envReadString(s string) string {
	val := os.Getenv(s)
	return val
}

func envReadBool(s string) bool {
	val := os.Getenv(s)
	if val == "true" || val == "1" {
		return true
	} else if val == "" || val == "false" || val == "0" {
		return false
	} else {
		panic(fmt.Errorf("failed to make boolean out of value: %s", val))
	}
}

func GetConfigDir() string {
	configDir := envReadString("CONFIG_PATH")
	if configDir == "" {
		configDir = GetAppRootDir() + "/config"
	}
	return configDir
}

func GetEnvFile() string {
	envFile := envReadString("ENV_FILE")
	if envFile == "" {
		envFile = filepath.Join(GetConfigDir(), ".env")
	}
	return envFile
}

func GetWorkerCount() int {
	workerCountStr := envReadString("POOL_WORKERS_COUNT")
	var workerCount int64
	if workerCountStr == "" {
		workerCount = int64(runtime.NumCPU() - 2)
	} else {
		var err error
		workerCount, err = strconv.ParseInt(workerCountStr, 10, 64)
		if err != nil {
			panic(err)
		}
	}
	return int(workerCount)
}

var appRoot string
func GetAppRootDir() string {
	if appRoot != "" {
		return appRoot
	}

	appRoot = envReadString("APP_ROOT")
	if appRoot == "" {
		wd, err := filepath.Abs(".")
		if err != nil {
			panic(err)
		}
		weblensIndex := strings.LastIndex(wd, "weblens")

		if weblensIndex == -1 {
			appRoot = "/app"
			log.Info.Println("APP_ROOT not set and could not be calculated, defaulting to", appRoot)
		} else {
			appRoot = wd[:weblensIndex+len("weblens")] + "/"
		}
	}
	return appRoot
}

func SetAppRoot(path string) {
	appRoot = path
}

func GetRouterIp() string {
	ip := envReadString("SERVER_IP")
	if ip == "" {
		log.Info.Println("SERVER_IP not provided, falling back to 0.0.0.0")
		return "0.0.0.0"
	} else {
		return ip
	}
}

func GetRouterPort() string {
	port := envReadString("SERVER_PORT")
	if port == "" {
		log.Info.Println("SERVER_PORT not provided, falling back to 8080")
		return "8080"
	} else {
		return port
	}
}

var mediaRoot string

func GetMediaRootPath() string {
	if mediaRoot != "" {
		return mediaRoot
	}

	path := envReadString("MEDIA_ROOT_PATH")
	if path == "" {
		path = "/media"
		log.Warning.Println("Did not find MEDIA_ROOT_PATH, assuming docker default of", path)
	}
	if path[len(path)-1:] != "/" {
		path = path + "/"
	}

	mediaRoot = path

	return mediaRoot
}

func GetExternalPaths() []string {
	return strings.Split(envReadString("EXTERNAL_PATHS"), " ")
}

func GetImgRecognitionUrl() string {
	return envReadString("IMG_RECOGNITION_URI")
}

var isDevMode *bool

func SetDevMode(devMode bool) {
	isDevMode = &devMode
}

// IsDevMode Enables debug logging and puts the router in development mode
func IsDevMode() bool {
	if isDevMode == nil {
		dev := envReadBool("DEV_MODE")
		isDevMode = &dev
	}
	return *isDevMode
}

// DetachUi Controls if we host UI comm on this server. UI can be hosted elsewhere and
// must proxy any /api/* requests back to this server
func DetachUi() bool {
	return envReadBool("DETATCH_UI")
}

var cachesPath string

func GetCacheRoot() string {
	if cachesPath == "" {
		cachesPath = envReadString("CACHES_PATH")
		if cachesPath == "" {
			cachesPath = "/cache"
			log.Warning.Println("Did not find CACHES_PATH, assuming docker default of", cachesPath)
		}
	}
	return cachesPath
}

// GetThumbsDir
// Returns the path of the directory for storing cached files. This includes photo thumbnails,
// temp uploaded files, and zip files.
func GetThumbsDir() string {
	cacheString := GetCacheRoot() + "/cache"
	_, err := os.Stat(cacheString)
	if err != nil {
		err = os.MkdirAll(cacheString, 0755)
		if err != nil {
			panic("CACHES_PATH provided, but the cache dir (`CACHES_PATH`/cache) does not exist and Weblens failed to create it")
		}
	}
	return cacheString
}

// GetTakeoutDir Takeout directory, stores zip files after creation
func GetTakeoutDir() string {
	takeoutString := GetCacheRoot() + "/takeout"
	_, err := os.Stat(takeoutString)
	if err != nil {
		err = os.MkdirAll(takeoutString, 0755)
		if err != nil {
			panic("CACHES_PATH provided, but the takeout dir (`CACHES_PATH`/takeout) does not exist and Weblens failed to create it")
		}
	}
	return takeoutString
}

func GetTmpDir() string {
	tmpString := GetCacheRoot() + "/tmp"
	_, err := os.Stat(tmpString)
	if err != nil {
		err = os.MkdirAll(tmpString, 0755)
		if err != nil {
			log.ShowErr(err)
			panic("CACHES_PATH provided, but the tmp dir (`CACHES_PATH`/tmp) does not exist and Weblens failed to create it")
		}
	}
	return tmpString
}

func GetMongoURI() string {
	mongoStr := envReadString("MONGODB_URI")
	if mongoStr == "" {
		mongoStr = "mongodb://localhost:27017"
		log.Warning.Println("MONGODB_URI not set, defaulting to", mongoStr)
	} else {
		log.Debug.Printf("Got MONGODB_URI: %s\n", mongoStr)
	}
	return mongoStr
}

func GetMongoDBName() string {
	mongoDBName := envReadString("MONGODB_NAME")
	if mongoDBName == "" {
		mongoDBName = "weblens"
	}
	return mongoDBName
}

func GetVideoConstBitrate() int {
	return 400000 * 2
}

var hostUrl string

func GetHostURL() string {
	if hostUrl == "" {
		hostUrl = envReadString("HOST_URL")
	}
	return hostUrl
}

var testMediaPath string

func GetTestMediaPath() string {
	if testMediaPath != "" {
		return testMediaPath
	}
	testMediaPath = envReadString("TEST_MEDIA_PATH")
	if testMediaPath == "" {
		testMediaPath = filepath.Join(GetAppRootDir(), "/images/testMedia")
		log.Warning.Printf("TEST_MEDIA_PATH not set, defaulting to %s", testMediaPath)
	}
	return testMediaPath
}

func ReadTypesConfig(target any) {
	typeJson, err := os.Open(filepath.Join(GetConfigDir(), "mediaType.json"))
	if err != nil {
		panic(err)
	}
	defer func(typeJson *os.File) {
		err := typeJson.Close()
		if err != nil {
			panic(err)
		}
	}(typeJson)

	typesBytes, err := io.ReadAll(typeJson)
	// marshMap := map[string]models.MediaType{}
	err = json.Unmarshal(typesBytes, target)
	if err != nil {
		panic(err)
	}
}

func GetCoreApiKey() string {
	return envReadString("CORE_API_KEY")
}