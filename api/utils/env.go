package util

import (
	"fmt"
	"os"
)

func EnvReadString(s string) (string) {
	val := os.Getenv(string(s))
	return val
}
func EnvReadBool(s string) (bool) {
	val := os.Getenv(string(s))
	if val == "true" || val == "1" {
		return true
	} else if val == "false" || val == "0" {
		return false
	} else {
		panic(fmt.Errorf("failed to make boolean out of value: %s", val))
	}
}