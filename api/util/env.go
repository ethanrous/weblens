package util

import (
	"fmt"
	"os"
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
		return "0.0.0.0"
	} else {
		return ip
	}
}

func GetRouterPort() string {
	port := envReadString("SERVER_PORT")
	if port == "" {
		return "8080"
	} else {
		return port
	}
}

func GetMediaRoot() string {
	return envReadString("MEDIA_ROOT_PATH")
}

func GetImgRecognitionUrl() string {
	return envReadString("IMG_RECOGNITION_URI")
}

func IsDevMode() bool {
	return envReadBool("DEV_MODE")
}

func DetachUi() bool {
	return envReadBool("DETATCH_UI")
}

func ShouldUseRedis() bool {
	return envReadBool("USE_REDIS")
}

func GetCacheDir() string {
	cacheString := envReadString("CACHES_PATH") + "/cache"
	_, err := os.Stat(cacheString)
	if err != nil {
		os.Mkdir(cacheString, 0755)
	}
	return cacheString
}

func GetTakeoutDir() string {
	takeoutString := envReadString("CACHES_PATH") + "/takeout"
	_, err := os.Stat(takeoutString)
	if err != nil {
		os.Mkdir(takeoutString, 0755)
	}
	return takeoutString
}

func GetTmpDir() string {
	tmpString := envReadString("CACHES_PATH") + "/tmp"
	_, err := os.Stat(tmpString)
	if err != nil {
		os.Mkdir(tmpString, 0755)
	}
	return tmpString
}

func GetMongoURI() string {
	return envReadString("MONGODB_URI")
}

func GetRedisUrl() string {
	return envReadString("REDIS_URL")
}

func GetLibRawPath() string {
	return envReadString("LIBRAW_PATH")
}
