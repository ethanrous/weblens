package internal

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	error2 "github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
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

func GetWorkerCount() int {
	workerCountStr := envReadString("POOL_WORKERS_COUNT")
	var workerCount int64
	if workerCountStr == "" {
		workerCount = int64(runtime.NumCPU() - 2)
	} else {
		var err error
		workerCount, err = strconv.ParseInt(workerCountStr, 10, 64)
		if err != nil {
			panic(error2.Wrap(err))
		}
	}
	return int(workerCount)
}

func GetAppRootDir() string {
	apiDir := envReadString("APP_ROOT")
	if apiDir == "" {
		apiDir = "/app"
		wlog.Info.Println("Api root directory not set, defaulting to", apiDir)
	}
	return apiDir
}

func GetRouterIp() string {
	ip := envReadString("SERVER_IP")
	if ip == "" {
		wlog.Info.Println("SERVER_IP not provided, falling back to 0.0.0.0")
		return "0.0.0.0"
	} else {
		return ip
	}
}

func GetRouterPort() string {
	port := envReadString("SERVER_PORT")
	if port == "" {
		wlog.Info.Println("SERVER_PORT not provided, falling back to 8080")
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
		wlog.Warning.Println("Did not find MEDIA_ROOT_PATH, assuming docker default of", path)
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

// IsDevMode Enables debug logging and puts the router in development mode
func IsDevMode() bool {
	if isDevMode == nil {
		dev := envReadBool("DEV_MODE")
		isDevMode = &dev
	}
	return *isDevMode
}

// DetachUi Controls if we host UI http on this server. UI can be hosted elsewhere and
// must proxy any /api/* requests back to this server
func DetachUi() bool {
	return envReadBool("DETATCH_UI")
}

var cachesPath string

func getCacheRoot() string {
	if cachesPath == "" {
		cachesPath = envReadString("CACHES_PATH")
		if cachesPath == "" {
			cachesPath = "/cache"
			wlog.Warning.Println("Did not find CACHES_PATH, assuming docker default of", cachesPath)
		}
	}
	return cachesPath
}

// GetCacheDir
// Returns the path of the directory for storing cached files. This includes photo thumbnails,
// temp uploaded files, and zip files.
func GetCacheDir() string {
	cacheString := getCacheRoot() + "/cache"
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
	takeoutString := getCacheRoot() + "/takeout"
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
	tmpString := getCacheRoot() + "/tmp"
	_, err := os.Stat(tmpString)
	if err != nil {
		err = os.MkdirAll(tmpString, 0755)
		if err != nil {
			wlog.ShowErr(err)
			panic("CACHES_PATH provided, but the tmp dir (`CACHES_PATH`/tmp) does not exist and Weblens failed to create it")
		}
	}
	return tmpString
}

func GetMongoURI() string {
	mongoStr := envReadString("MONGODB_URI")
	if mongoStr == "" {
		wlog.Error.Panicf("MONGODB_URI not set! MongoDB is required to use Weblens. Docs for mongo connection strings are here:\nhttps://www.mongodb.com/docs/manual/reference/connection-string/")
	}
	wlog.Debug.Printf("Using MONGODB_URI: %s\n", mongoStr)
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
