package util

import (
	"fmt"
	"os"
)

func envReadString(s string) (string) {
	val := os.Getenv(string(s))
	return val
}
func envReadBool(s string) (bool) {
	val := os.Getenv(string(s))
	if val == "true" || val == "1" {
		return true
	} else if val == "false" || val == "0" {
		return false
	} else {
		panic(fmt.Errorf("failed to make boolean out of value: %s", val))
	}
}

func GetMediaRoot() (string) {
	return envReadString("MEDIA_ROOT_PATH")
}

func IsDevMode() (bool) {
	return envReadBool("DEV_MODE")
}

func ShouldUseRedis() (bool) {
	return envReadBool("USE_REDIS")
}

func ShowDebug() (bool) {
	return envReadBool("SHOW_DEBUG")
}

func GetTakeoutDir() (string) {
	takeoutString := envReadString("CACHES_PATH") + "/takeout"
	_, err := os.Stat(takeoutString)
	if err != nil {
		os.Mkdir(takeoutString, 0755)
	}
	return takeoutString
}

func GetTmpDir() (string) {
	tmpString := envReadString("CACHES_PATH") + "/tmp"
	_, err := os.Stat(tmpString)
	if err != nil {
		os.Mkdir(tmpString, 0755)
	}
	return tmpString
}

func GetTrashDir() (string) {
	trashString := envReadString("CACHES_PATH") + "/trash"
	_, err := os.Stat(trashString)
	if err != nil {
		os.Mkdir(trashString, 0755)
	}
	return trashString
}

func GetMongoURI() (string) {
	return envReadString("MONGODB_URI")
}

func GetRedisUrl() (string) {
	return envReadString("REDIS_URL")
}

func GetLibRawPath() (string) {
	return envReadString("LIBRAW_PATH")
}