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

func getMediaRoot() (string) {
	return envReadString("MEDIA_ROOT_PATH")
}

func GetTrashDir() (string) {
	return envReadString("TRASH_PATH")
}

func GetMongoURI() (string) {
	return envReadString("MONGODB_URI")
}