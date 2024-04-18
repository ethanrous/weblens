package util

import (
	"fmt"
	"os"
	"strings"
)

func envReadString(s string) string {
	val := os.Getenv(string(s))
	return val
}
func envReadBool(s string) bool {
	val := os.Getenv(string(s))
	if val == "true" || val == "1" {
		return true
	} else if val == "" || val == "false" || val == "0" {
		return false
	} else {
		panic(fmt.Errorf("failed to make boolean out of value: %s", val))
	}
}

func GetConfigDir() string {
	configDir := envReadString("CONFIG_DIR")
	if configDir == "" {
		configDir = "/app/config"
		Info.Println("Config directory not set, using", configDir)
	}
	return configDir
}

func GetRouterIp() string {
	ip := envReadString("SERVER_IP")
	if ip == "" {
		Warning.Println("SERVER_IP not provided, falling back to 0.0.0.0")
		return "0.0.0.0"
	} else {
		Debug.Printf("Using SERVER_IP: %s\n", ip)
		return ip
	}
}

func GetRouterPort() string {
	port := envReadString("SERVER_PORT")
	if port == "" {
		Warning.Println("SERVER_PORT not provided, falling back to 8080")
		return "8080"
	} else {
		Debug.Printf("Using SERVER_PORT: %s\n", port)
		return port
	}
}

func GetMediaRootPath() string {
	path := envReadString("MEDIA_ROOT_PATH")
	if path == "" {
		panic("MEDIA_ROOT_PATH not set! This is required as it is the primary storage location for weblens files")
	}
	return path
}

func GetExternalPaths() []string {
	return strings.Split(envReadString("EXTERNAL_PATHS"), " ")
}

func GetImgRecognitionUrl() string {
	return envReadString("IMG_RECOGNITION_URI")
}

// Enables debug logging and puts the router in development mode
func IsDevMode() bool {
	return envReadBool("DEV_MODE")
}

// Controls if we host UI routes on this server. UI can be hosted elsewhere and
// must proxy any /api/* requests back to this server
func DetachUi() bool {
	return envReadBool("DETATCH_UI")
}

// Enable use of redis on this server. If true, "REDIS_URL" must also be set
func ShouldUseRedis() bool {
	return envReadBool("USE_REDIS")
}

func GetCachesPath() string {
	return envReadString("CACHES_PATH")
}

// Directory for storing cached files. This includes photo thumbnails,
// temp uploaded files, and zip files.
func GetCacheDir() string {
	cacheString := envReadString("CACHES_PATH") + "/cache"
	_, err := os.Stat(cacheString)
	if err != nil {
		err = os.MkdirAll(cacheString, 0755)
		if err != nil {
			panic("CACHES_PATH provided, but the cache dir (`CACHES_PATH`/cache) does not exist and Weblens failed to create it")
		}
	}
	return cacheString
}

// Takeout directory, stores zip files after creation
func GetTakeoutDir() string {
	takeoutString := envReadString("CACHES_PATH") + "/takeout"
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
	caches := envReadString("CACHES_PATH")
	if caches == "" {
		panic("CACHES_PATH not provided")
	}
	tmpString := caches + "/tmp"
	_, err := os.Stat(tmpString)
	if err != nil {
		err = os.MkdirAll(tmpString, 0755)
		if err != nil {
			ShowErr(err)
			panic("CACHES_PATH provided, but the tmp dir (`CACHES_PATH`/tmp) does not exist and Weblens failed to create it")
		}
	}
	return tmpString
}

func GetMongoURI() string {
	mongoStr := envReadString("MONGODB_URI")
	if mongoStr == "" {
		Error.Panicf("MONGODB_URI not set! MongoDB is required to use Weblens. Docs for mongo connection strings are here:\nhttps://www.mongodb.com/docs/manual/reference/connection-string/")
	}
	Info.Printf("Using MONGODB_URI: %s\n", mongoStr)
	return mongoStr
}

func GetMongoDBName() string {
	mongoDBName := envReadString("MONGODB_NAME")
	if mongoDBName == "" {
		mongoDBName = "weblens"
	}
	return mongoDBName
}

func GetRedisUrl() string {
	return envReadString("REDIS_URL")
}
