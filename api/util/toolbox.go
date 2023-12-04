package util

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func FailOnError(err error, format string, fmtArgs... any) {
	msg := fmt.Sprintf(format, fmtArgs...)
	if err != nil {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			var trace []byte
			runtime.Stack(trace, false)
			panic(fmt.Errorf("error from %s:%d %s: %s\n%s", file, line, msg, err, string(trace)))
		} else {
			Error.Panicf("Failed to get caller information while parsing this error:\n%s: %s", msg, err)
		}
	}
}

func DisplayError(err error, extras... string) {
	msg := strings.Join(extras, " ")
	if err != nil {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			var trace []byte
			runtime.Stack(trace, false)
			Error.Printf("Error from %s:%d %s: %s\n%s", file, line, msg, err, string(trace))
		} else {
			Error.Printf("Failed to get caller information while parsing this error:\n%s: %s", msg, err)
		}
	}
}

func Warn(err error, extras... string) {
	msg := strings.Join(extras, " ")
	if err != nil {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			var trace []byte
			runtime.Stack(trace, false)
			Warning.Printf("Warning from %s:%d %s: %s\n%s", file, line, msg, err, string(trace))
		} else {
			Warning.Printf("Failed to get caller information while parsing this error:\n%s: %s", msg, err)
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

// Set charLimit to 0 to disable
func HashOfString(charLimit int, dataToHash... string) string {
	h := sha256.New()

	for _, s := range dataToHash {
		h.Write([]byte(s))
	}

	hash := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if charLimit != 0 {
		return hash[:charLimit]
	} else {
		return hash
	}
}

func StructFromMap(inputMap map[string]any, target any) error {
	jsonString, err := json.Marshal(inputMap)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonString, target)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func LazyStackTrace() {
	Debug.Println("\n----- Lazy Stack Trace (most recent last, showing 5) -----")
	for i := 5; i > 0; i-- {
		_, file, line, _ := runtime.Caller(i)
		Debug.Printf("%s:%d\n", file, line)
	}
}

func identifyPanic() string {
	var name, file string
	var line int
	var pc [16]uintptr

	n := runtime.Callers(3, pc[:])
	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line = fn.FileLine(pc)
		name = fn.Name()
		if !strings.HasPrefix(name, "runtime.") || strings.HasPrefix(name, "/opt/homebrew") {
			break
		}
	}

	if file != "" {
		return fmt.Sprintf("%v:%v", filepath.Base(file), line)
	}

	return fmt.Sprintf("pc:%x", pc)
}

func RecoverPanic(preText ...any) {
	r := recover()
	if r == nil {
		return
	}
	var formatString string
	var rest []any
	if len(preText) != 0 {
		formatString = preText[0].(string)
		rest = preText[1:]
	} else {
		formatString = ""
		rest = []any{}
	}
	ErrorCatcher.Println(fmt.Sprintf(formatString, rest...), identifyPanic(), r)
}

func Map[T, V any](ts []T, fn func(T) V) []V {
    result := make([]V, len(ts))
    for i, t := range ts {
        result[i] = fn(t)
    }
    return result
}

func Filter[T any](ts []T, fn func(T) bool) []T {
    var result []T
    for _, t := range ts {
        if fn(t) {
			result = append(result, t)
		}
    }
    return result
}