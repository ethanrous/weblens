package util

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func BoolFromString(input string, emptyIsFalse bool) bool {
	if input == "1" || input == "true" {
		return true
	} else if input == "0" || input == "false" {
		return false
	} else if input == "" {
		return !emptyIsFalse
	} else {
		panic(fmt.Errorf("unexpected boolean string. Unable to determine truthiness of %s", input))
	}
}

func IntFromBool( input bool ) int {
	if input {
		return 1
	} else {
		return 0
	}
}

func SliceRemove(s []any, i int) []any {
    s[i] = s[len(s)-1]
    return s[:len(s)-1]
}

func FailOnError(err error, msg string) {
	if err != nil {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			Error.Panicf("Erorr from %s:%d %s: %s", file, line, msg, err)
		} else {
			Error.Panicf("Failed to get caller information while parsing this error:\n%s: %s", msg, err)
		}
	}
}

func FailOnNoError(err error, msg string) {
	if err == nil {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			Error.Panicf("Expected error from %s:%d %s", file, line, msg)
		} else {
			Error.Panicf("Failed to get caller information while expecting error:\n%s", msg)
		}
	}
}

func DirSize(path string) (int64, error) {
    var size int64
    err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() {
            size += info.Size()
        }
        return err
    })
    return size, err
}

func bToMb(b uint64) uint64 {
    return b / 1024 / 1024
}

func PrintMemUsage() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(mem.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(mem.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(mem.Sys))
	fmt.Printf("\tNumGC = %v\n", mem.NumGC)
}