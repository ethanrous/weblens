package util

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
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

func ErrTrace(err error, extras ...string) {
	if err != nil {
		_, file, line, ok := runtime.Caller(1)
		msg := strings.Join(extras, " ")
		if ok {
			ErrorCatcher.Printf(
				"%s:%d %s: %s\n----- STACK FOR ERROR ABOVE -----\n%s", file, line, msg, err, debug.Stack(),
			)
		} else {
			Error.Printf("Failed to get caller information while parsing this error:\n%s: %s", msg, err)
		}
	}
}

func ShowErr(err error, extras ...string) {
	if err != nil {
		_, file, line, ok := runtime.Caller(1)
		msg := strings.Join(extras, " ")
		if ok {
			ErrorCatcher.Printf("%s:%d %s: %s", file, line, msg, err)
		} else {
			Error.Printf("Failed to get caller information while parsing this error:\n%s: %s", msg, err)
		}
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

type WeblensHash struct {
	hash hash.Hash
}

func NewWeblensHash() *WeblensHash {
	return &WeblensHash{hash: sha256.New()}
}

func (h *WeblensHash) Add(data []byte) error {
	_, err := h.hash.Write(data)
	return err
}

func (h *WeblensHash) Done(len int) string {
	return base64.URLEncoding.EncodeToString(h.hash.Sum(nil))[:len]
}

// GlobbyHash Set charLimit to 0 to disable
func GlobbyHash(charLimit int, dataToHash ...any) string {
	h := NewWeblensHash()

	s := fmt.Sprint(dataToHash...)
	h.Add([]byte(s))

	if charLimit != 0 && charLimit < 16 {
		return h.Done(charLimit)
	} else {
		return h.Done(16)
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

func RecoverPanic(preText string) {
	r := recover()
	if r == nil {
		return
	}

	ErrorCatcher.Println(preText, IdentifyPanic(), r)
}

// Yoink See Banish. Yoink is the same as Banish, but returns the value at i
// in addition to the shortened slice.
func Yoink[T any](s []T, i int) ([]T, T) {
	t := s[i]
	n := append(s[:i], s[i+1:]...)
	return n, t
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

func OnlyUnique[T comparable](s []T) (rs []T) {
	tmpMap := make(map[T]bool, len(s))
	for _, t := range s {
		tmpMap[t] = true
	}
	rs = MapToKeys(tmpMap)
	return
}

// Banish removes the element at index, i, from the slice, s, in place
//
// Banish returns a slice of length len(s) - 1
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

func InsertFunc[S ~[]T, T any](ts S, t T, cmp func(a T, b T) int) S {
	i, _ := slices.BinarySearchFunc(ts, t, cmp) // find slot
	return slices.Insert(ts, i, t)
}

func Diff[T comparable](s1 []T, s2 []T) []T {
	if len(s1) < len(s2) {
		s1, s2 = s2, s1
	}
	var res []T
	for _, t := range s1 {
		if !slices.Contains(s2, t) {
			res = append(res, t)
		}
	}

	return res
}

func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

func Each[T any](ts []T, fn func(T)) {
	if ts == nil {
		return
	}
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

func Filter[S ~[]T, T any](ts S, fn func(t T) bool) []T {
	var result []T
	for _, t := range ts {
		if fn(t) {
			result = append(result, t)
		}
	}
	return result
}

func FilterMap[T, V any](ts []T, fn func(T) (V, bool)) []V {
	var result []V
	for _, t := range ts {
		res, y := fn(t)
		if y {
			result = append(result, res)
		}
	}
	return result
}

func Reduce[T, A any](ts []T, fn func(T, A) A, acc A) A {
	for _, t := range ts {
		acc = fn(t, acc)
	}
	return acc
}

// SliceConvert Perform type assertion on slice
func SliceConvert[V, T any](ts []T) []V {
	vs := make([]V, len(ts))
	if len(ts) == 0 {
		return vs
	}
	for i := range ts {
		vs[i] = any(ts[i]).(V)
	}

	return vs
}

type lap struct {
	tag  string
	time time.Time
}

type Stopwatch interface {
	Lap(tag ...any)
	Stop() time.Duration
	PrintResults(firstLapIsStart bool)
	GetTotalTime(bool) time.Duration
}

type sw struct {
	name  string
	start time.Time
	laps  []lap
	stop  time.Time
}

type prod_sw bool

func (prod_sw) Stop() (t time.Duration)         { return }
func (prod_sw) Lap(tag ...any)                  {}
func (prod_sw) PrintResults(bool)               {}
func (prod_sw) GetTotalTime(bool) time.Duration { return time.Millisecond * 0 }

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
	// if l.tag != "" {
	// 	Debug.Println(l.tag)
	// }
}

func (s *sw) GetTotalTime(firstLapIsStart bool) time.Duration {
	var start time.Time
	var end time.Time

	if s.stop.Unix() < 0 {
		end = time.Now()
	} else {
		end = s.stop
	}

	if firstLapIsStart && len(s.laps) > 0 {
		start = s.laps[0].time
	} else {
		start = s.start
	}

	return end.Sub(start)
}

func (s *sw) PrintResults(firstLapIsStart bool) {
	if s.stop.Unix() < 0 {
		Error.Println("Stopwatch cannot provide results before being stopped")
		return
	}

	var res = fmt.Sprintf("--- %s Stopwatch ---", s.name)

	var startTime time.Time
	if firstLapIsStart {
		if len(s.laps) <= 1 {
			return
		}
		startTime = s.laps[0].time
	} else {
		startTime = s.start
	}

	if len(s.laps) != 0 {
		longest := len(slices.MaxFunc(s.laps, func(a, b lap) int { return len(a.tag) - len(b.tag) }).tag)
		lapFmt := fmt.Sprintf("\t%%-%ds %%-15s (%%s since start -- %%s since creation)", longest+5)
		for i, l := range s.laps {
			var sinceLast time.Duration
			if i != 0 {
				sinceLast = s.laps[i].time.Sub(s.laps[i-1].time)
			} else {
				sinceLast = s.laps[i].time.Sub(s.start)
			}

			if l.tag != "" {
				res = fmt.Sprintf(
					"%s\n%s", res, fmt.Sprintf(lapFmt, l.tag, sinceLast, l.time.Sub(startTime), l.time.Sub(s.start)),
				)
			}
		}
	}

	fmt.Printf("%s\n%s\n", res, fmt.Sprintf("Stopped at %s", s.stop.Sub(startTime)))
}

// OracleReader is almost exactly like io.ReadAll, but if we know how long the content is,
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