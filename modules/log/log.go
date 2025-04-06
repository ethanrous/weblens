package log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

// Deprecated: Use json logger instead.

var OpenSearchClient *opensearch.Client

var projectPackagePrefix string

func init() {

	_, filename, _, _ := runtime.Caller(0)
	projectPackagePrefix = strings.TrimSuffix(filename, "modules/log/log.go")
	if projectPackagePrefix == filename {
		// in case the source code file is moved, we can not trim the suffix, the code above should also be updated.
		panic("weblens logger unable to detect correct package prefix, please update file: " + filename)
	}

	var err error
	osUrl := os.Getenv("OPENSEARCH_URL")
	osUser := os.Getenv("OPENSEARCH_USER")
	osPass := os.Getenv("OPENSEARCH_PASSWORD")

	if osUrl != "" && osUser != "" && osPass != "" {
		OpenSearchClient, err = NewOpenSearchClient(osUrl, osUser, osPass)
	}
	if err != nil {
		panic(err)
	}

	zerolog.TimestampFieldName = "@timestamp"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.ErrorStackFieldName = "traceback"
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

var logger *zerolog.Logger

func NopLogger() *zerolog.Logger {
	// nop := zerolog.Nop()
	// return &nop

	if logger != nil {
		return logger
	}
	return NewZeroLogger()

}
func NewZeroLogger() *zerolog.Logger {
	if logger != nil {
		return logger
	}

	var localLogger io.Writer
	if os.Getenv("LOG_FORMAT") == "dev" {
		localLogger = WLConsoleLogger{}
	} else {
		localLogger = os.Stdout
	}

	writers := []io.Writer{localLogger}

	osIndex := os.Getenv("OPENSEARCH_INDEX")
	if OpenSearchClient != nil {
		oLog := NewOpensearchLogger(OpenSearchClient, osIndex)
		writers = append(writers, oLog)
	}

	wl_version := os.Getenv("WEBLENS_BUILD_VERSION")
	if wl_version == "" {
		wl_version = "unknown"

		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			for _, v := range buildInfo.Settings {
				if v.Key == "vcs.revision" {
					wl_version = v.Value
					break
				}
			}
		}
	}

	multi := zerolog.MultiLevelWriter(writers...)
	log := zerolog.New(multi).Level(zerolog.DebugLevel).With().Timestamp().Caller().Str("weblens_build_version", wl_version).Logger()
	logger = &log
	zlog.Logger = log

	return &log
}

type WLConsoleLogger struct {
}

func (l WLConsoleLogger) Write(p []byte) (n int, err error) {
	var target map[string]any
	err = json.Unmarshal(p, &target)
	if err != nil {
		return
	}

	var caller string
	pc, filename, line, ok := runtime.Caller(6)
	if ok {
		fn := runtime.FuncForPC(pc)
		var fnName string
		if fn != nil {
			fnName = fn.Name()
			fnName = strings.ReplaceAll(fnName, "[...]", "") + "()" // generic function names are "foo[...]"

			slash := strings.LastIndex(fnName, "/") + 1
			dot := strings.LastIndex(fnName[slash:], ".") + 1

			fnName = fnName[slash+dot:]
		}
		filename = strings.TrimPrefix(filename, projectPackagePrefix)
		caller = fmt.Sprintf("%s:%d:%s", filename, line, fnName)
	}

	level, _ := target["level"].(string)
	msgErr, _ := target["error"].(string)
	logMsg, _ := target["message"].(string)
	traceback, _ := target["traceback"].([]any)

	ignoredKeys := []string{"level", "error", "message", "weblens_build_version", "caller", "@timestamp", "traceback", "instance"}

	extras := ""
	for k, v := range target {
		if slices.Contains(ignoredKeys, k) {
			continue
		}
		extras += fmt.Sprintf("%s%s%s: %v ", BLUE, k, RESET, v)
	}
	if extras != "" {
		logMsg += "\n\t" + extras
	}

	var stackStr string
	if len(traceback) != 0 {
		stackStr = "\n"
		for _, block := range traceback {
			blockMap, ok := block.(map[string]any)
			if !ok {
				continue
			}

			source, okSource := blockMap["source"].(string)
			line, _ := blockMap["line"].(string)
			function, _ := blockMap["func"].(string)
			if okSource {
				source = strings.TrimPrefix(source, projectPackagePrefix)
				fileAndLine := source + ":" + line
				function += "()"
				stackStr += fmt.Sprintf("\t%s%-25s %s%30s\n", BLUE, fileAndLine, RESET, function)
			}
		}
	}

	var levelColor string
	switch level {
	case "trace":
		levelColor = RESET
	case "debug":
		levelColor = ORANGE
	case "info":
		levelColor = BLUE
	case "warn":
		levelColor = YELLOW
	case "error", "fatal":
		levelColor = RED
	}

	timeStr := time.Now().Format(time.DateTime)
	msg := fmt.Sprintf("%s%s %s%s %s[%s] %s%s %s%s%s%s\n", ORANGE, timeStr, BLUE, caller, levelColor, level, RESET, logMsg, RED, msgErr, stackStr, RESET)
	_, err = os.Stdout.Write([]byte(msg))
	if err != nil {
		return
	}
	return len(p), nil
}
