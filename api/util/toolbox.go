package util

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"time"
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

func FailOnError(err error, format string, fmtArgs ...any) {
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

func DisplayError(err error, extras ...string) {
	if err != nil {
		_, file, line, ok := runtime.Caller(1)
		msg := strings.Join(extras, " ")
		if ok {
			ErrorCatcher.Printf("%s:%d %s: %s\n%s", file, line, msg, err, debug.Stack())
		} else {
			Error.Printf("Failed to get caller information while parsing this error:\n%s: %s", msg, err)
		}
	}
}

func Warn(err error, extras ...string) {
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

func PrintMemUsage(contextTag string) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	Debug.Println("Memory dump: ", contextTag)
	fmt.Printf("Alloc = %v MiB", bToMb(mem.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(mem.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(mem.Sys))
	fmt.Printf("\tNumGC = %v\n", mem.NumGC)
}

// Set charLimit to 0 to disable
func GlobbyHash(charLimit int, dataToHash ...any) string {
	h := sha256.New()

	s := fmt.Sprint(dataToHash...)
	h.Write([]byte(s))

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

func IdentifyPanic() string {
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
	ErrorCatcher.Println(fmt.Sprintf(formatString, Map(rest, func(a any) string { return a.(string) })), IdentifyPanic(), r)
}

// See Banish. Yoink is the same as Banish, but returns the value at i
// in addition to the shortened slice
func Yoink[T any](s []T, i int) ([]T, T) {
	y := s[i]
	s[i] = s[len(s)-1]
	return s[:len(s)-1], y
}

func YoinkFunc[T any](s []T, fn func(f T) bool) (rs []T, rt T, re bool) {
	for i, t := range s {
		if fn(t) {
			rs, rt = Yoink(s, i)
			re = true
			return
		}
	}
	rs = s
	return
}

// Banish removes the element at index, i, from the slice, s, in place and in constant time.
//
// Banish returns a slice of length len(s) - 1. The order of s will be modified
func Banish[T any](s []T, i int) []T {
	s, _ = Yoink(s, i)
	return s
}

func AddToSet[T comparable](set []T, add []T) []T {
	for _, a := range add {
		if !slices.Contains(set, a) {
			set = append(set, a)
		}
	}
	return set
}

func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

func Each[T any](ts []T, fn func(T)) {
	for _, t := range ts {
		fn(t)
	}
}

func Find[T any](ts []T, fn func(T) bool) int {
	for i, t := range ts {
		if fn(t) {
			return i
		}
	}
	return -1
}

func MapToSliceMutate[T comparable, X, V any](tMap map[T]X, fn func(T, X) V) []V {
	result := make([]V, len(tMap))
	counter := 0
	for t, x := range tMap {
		result[counter] = fn(t, x)
		counter++
	}
	return result
}

// Takes a generic map and returns a slice of the values
func MapToSlicePure[T comparable, V any](tMap map[T]V) []V {
	result := make([]V, len(tMap))
	counter := 0
	for _, v := range tMap {
		result[counter] = v
		counter++
	}
	return result
}

func MapToKeys[T comparable, V any](tMap map[T]V) []T {
	result := make([]T, len(tMap))
	counter := 0
	for t := range tMap {
		result[counter] = t
		counter++
	}
	return result
}

func Filter[T any](ts []T, fn func(T) bool) []T {
	var result []T = []T{}
	for _, t := range ts {
		if fn(t) {
			result = append(result, t)
		}
	}
	return result
}

func FilterMap[T, V any](ts []T, fn func(T) (V, bool)) []V {
	var result []V = []V{}
	for _, t := range ts {
		res, y := fn(t)
		if y {
			result = append(result, res)
		}
	}
	return result
}

type lap struct {
	tag  string
	time time.Time
}

type Stopwatch interface {
	Lap(tag ...any)
	Stop() time.Duration
	PrintResults()
}

type sw struct {
	name  string
	start time.Time
	laps  []lap
	stop  time.Time
}

type prod_sw bool

func (prod_sw) Stop() (t time.Duration) { return }
func (prod_sw) Lap(tag ...any)          {}
func (prod_sw) PrintResults()           {}

func NewStopwatch(name string) Stopwatch {
	if IsDevMode() {
		return &sw{name: name, start: time.Now()}
	}
	return prod_sw(false)
}

func (s *sw) Stop() time.Duration {
	s.stop = time.Now()
	return s.stop.Sub(s.start)
}

func (s *sw) Lap(tag ...any) {
	l := lap{
		tag:  fmt.Sprint(tag...),
		time: time.Now(),
	}
	s.laps = append(s.laps, l)
}

func (s *sw) PrintResults() {
	if s.stop.Unix() < 0 {
		Error.Println("Stopwatch cannot provide results before being stopped")
		return
	}

	var res string = fmt.Sprintf("--- %s Stopwatch ---", s.name)

	if len(s.laps) != 0 {
		longest := len(slices.MaxFunc(s.laps, func(a, b lap) int { return len(a.tag) - len(b.tag) }).tag)
		lapFmt := fmt.Sprintf("\t%%-%ds %%-15s (%%s since creation)", longest+5)
		for i, l := range s.laps {
			var sinceLast time.Duration
			if i != 0 {
				sinceLast = s.laps[i].time.Sub(s.laps[i-1].time)
			} else {
				sinceLast = s.laps[i].time.Sub(s.start)
			}

			if l.tag != "" {
				res = fmt.Sprintf("%s\n%s", res, fmt.Sprintf(lapFmt, l.tag, sinceLast, l.time.Sub(s.start)))
			}
		}
	}

	fmt.Printf("%s\n%s\n", res, fmt.Sprintf("Stopped at %s", s.stop.Sub(s.start)))
	// fmt.Println(res)
}

// Almost exactly like io.ReadAll, but if we know how long the content is,
// we can allocate the whole array up front, saving a bit of time
func OracleReader(r io.Reader, readerSize int64) ([]byte, error) {
	b := make([]byte, 0, readerSize)
	for {
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}

		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
	}
}
