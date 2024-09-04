package env

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/ethrousseau/weblens/internal/log"
	"github.com/joho/godotenv"
)

var envLock sync.RWMutex

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

var envRead = false

func ReadEnv() {
	envLock.Lock()
	defer envLock.Unlock()

	if envRead {
		return
	}

	err := godotenv.Load(GetEnvFile())
	if err != nil {
		log.Warning.Println("Failed to load env file")
	}
	envRead = true
}

var configData map[string]map[string]any

func ReadConfig(configPath, configName string) (map[string]any, error) {
	envLock.Lock()
	defer envLock.Unlock()
	if configData != nil {
		return configData[configName], nil
	}

	if configPath == "" {
		configDir := os.Getenv("CONFIG_PATH")
		configPath = filepath.Join(configDir, "config.json")
	}

	bs, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config map[string]map[string]any
	err = json.Unmarshal(bs, &config)
	if err != nil {
		return nil, err
	}

	configData = config
	return configData[configName], nil
}

func GetEnvFile() string {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		panic("ENV_FILE not set")
	}
	return envFile
}

func GetWorkerCount() int {
	config, err := ReadConfig("", os.Getenv("CONFIG_NAME"))
	if err == nil {
		countStr := config["poolWorkerCount"]
		if countStr != nil {
			count, err := strconv.ParseInt(countStr.(string), 10, 64)
			if err == nil {
				return int(count)
			}
		}
	}

	return runtime.NumCPU() - 2
}

func GetAppRootDir() string {
	rootPath := filepath.Dir(os.Getenv("CONFIG_PATH"))
	return rootPath
}

func GetRouterPort() string {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		log.Info.Println("SERVER_PORT not provided, falling back to 8080")
		return "8080"
	} else {
		return port
	}
}

func GetExternalPaths() []string {
	return strings.Split(os.Getenv("EXTERNAL_PATHS"), " ")
}

func GetImgRecognitionUrl() string {
	return os.Getenv("IMG_RECOGNITION_URI")
}

// IsDevMode Enables debug logging and puts the router in development mode
func IsDevMode() bool {
	config, err := ReadConfig("", os.Getenv("CONFIG_NAME"))
	if err != nil {
		return false
	}

	return config["logLevel"].(string) == "debug"
}

// DetachUi Controls if we host UI comm on this server. UI can be hosted elsewhere and
// must proxy any /api/* requests back to this server
func DetachUi() bool {
	return envReadBool("DETATCH_UI")
}

var cachesPath string

func GetCacheRoot() string {
	if cachesPath == "" {
		cachesPath = os.Getenv("CACHES_PATH")
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
	envLock.Lock()
	defer envLock.Unlock()
	mongoStr := os.Getenv("MONGODB_URI")
	if mongoStr == "" {
		mongoStr = "mongodb://localhost:27017"
		log.Warning.Println("MONGODB_URI not set, defaulting to", mongoStr)
	} else {
		log.Debug.Printf("Got MONGODB_URI: %s\n", mongoStr)
	}
	return mongoStr
}

func GetMongoDBName() string {
	config, err := ReadConfig("", os.Getenv("CONFIG_NAME"))
	if err != nil {
		panic(err)
	}
	return config["mongodbName"].(string)
}

var hostUrl string

func GetHostURL() string {
	if hostUrl == "" {
		hostUrl = os.Getenv("HOST_URL")
	}
	return hostUrl
}

var testMediaPath string

func GetTestMediaPath() string {
	envLock.Lock()
	defer envLock.Unlock()
	if testMediaPath != "" {
		return testMediaPath
	}

	testMediaPath = os.Getenv("TEST_MEDIA_PATH")
	if testMediaPath == "" {
		envLock.Unlock()
		testMediaPath = filepath.Join(GetAppRootDir(), "/images/testMedia")
		envLock.Lock()
		log.Warning.Printf("TEST_MEDIA_PATH not set, defaulting to %s", testMediaPath)
	}
	return testMediaPath
}

func ReadTypesConfig(target any) error {
	typeJson, err := os.Open(filepath.Join(os.Getenv("CONFIG_PATH"), "mediaType.json"))
	if err != nil {
		return err
	}
	defer func(typeJson *os.File) {
		err = typeJson.Close()
		if err != nil {
			log.ErrTrace(err)
		}
	}(typeJson)

	typesBytes, err := io.ReadAll(typeJson)
	err = json.Unmarshal(typesBytes, &target)
	if err != nil {
		return err
	}

	return nil
}

func GetCoreApiKey() string {
	return os.Getenv("CORE_API_KEY")
}

func GetMediaRoot() string {
	mediaRoot := os.Getenv("MEDIA_ROOT")
	if mediaRoot != "" {
		return mediaRoot
	}

	config, err := ReadConfig("", os.Getenv("CONFIG_PATH"))
	if err != nil {
		panic(err)
	}

	mediaRoot = config["mediaRoot"].(string)
	if mediaRoot[0] == '.' {
		mediaRoot = mediaRoot[1:]
		mediaRoot = filepath.Join(GetAppRootDir(), mediaRoot)
	}

	return mediaRoot
}

func GetCachesRoot() string {
	cachesRoot := os.Getenv("CACHES_ROOT")
	if cachesRoot != "" {
		return cachesRoot
	}

	config, err := ReadConfig("", os.Getenv("CONFIG_PATH"))
	if err != nil {
		panic(err)
	}

	cachesRoot = config["cachesRoot"].(string)
	if cachesRoot[0] == '.' {
		cachesRoot = cachesRoot[1:]
		cachesRoot = filepath.Join(GetAppRootDir(), cachesRoot)
	}

	return cachesRoot
}