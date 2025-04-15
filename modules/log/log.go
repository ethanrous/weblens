package log

import (
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

	"github.com/ethanrous/weblens/modules/config"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

// Deprecated: Use json logger instead.

var OpenSearchClient *opensearch.Client

var projectPrefix string

func init() {

	_, filename, _, _ := runtime.Caller(0)
	projectPrefix = strings.TrimSuffix(filename, "modules/log/log.go")
	if projectPrefix == filename {
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
	zerolog.ErrorStackMarshaler = MarshalStack
	zerolog.ErrorStackFieldName = "traceback"
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

type ErrUnwrapHook struct{}

func (h ErrUnwrapHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	fmt.Println("ErrUnwrapHook", e, level, msg)
}

var logger zerolog.Logger = zerolog.Nop()

func NopLogger() zerolog.Logger {
	nop := zerolog.Nop()
	return nop

	// return NewZeroLogger()
}

func NewZeroLogger() zerolog.Logger {
	if logger.GetLevel() != zerolog.Disabled {
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

	logLevel := config.GetConfig().LogLevel
	zerolog.SetGlobalLevel(logLevel)

	multi := zerolog.MultiLevelWriter(writers...)
	log := zerolog.New(multi).Level(logLevel).With().Timestamp().Caller().Str("weblens_build_version", wl_version).Logger()
	logger = log
	zlog.Logger = log

	log.Info().Msgf("Weblens logger initialized [%s]", log.GetLevel())
	log.Trace().Msgf("Weblens logger initialized [%s]", logLevel.String())

	return log
}

type WLConsoleLogger struct{}

func (l WLConsoleLogger) Write(p []byte) (n int, err error) {
	var target map[string]any
	err = json.Unmarshal(p, &target)
	if err != nil {
		return
	}

	callerI, ok := target["caller"]
	caller := "[NO CALLER]"
	if ok {
		caller = callerI.(string)
		caller = strings.TrimPrefix(caller, projectPrefix)
	}

	level, _ := target["level"].(string)
	msgErr, _ := target["error"].(string)
	logMsg, _ := target["message"].(string)
	traceback, _ := target["traceback"].([]any)

	if level == zerolog.TraceLevel.String() {
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

	}

	var stackStr string
	if len(traceback) != 0 {
		stackStr = "\n"
		for _, block := range traceback {
			blockMap, ok := block.(map[string]any)
			if !ok {
				continue
			}

			function, _ := blockMap["func"].(string)
			if strings.HasSuffix(function, "ServeHTTP") {
				break
			}

			source, okSource := blockMap["source"].(string)
			line, _ := blockMap["line"].(string)
			if okSource {
				if strings.HasPrefix(source, projectPrefix) {
					source = strings.TrimPrefix(source, projectPrefix)
				} else {
					source = ".../" + filepath.Base(source)
				}
				fileAndLine := source + ":" + line
				function += "()"
				stackStr += fmt.Sprintf("\t%s%-30s %s%30s\n", BLUE, fileAndLine, RESET, function)
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
