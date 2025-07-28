package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

var OpenSearchClient *opensearch.Client

var projectPrefix string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	projectPrefix = strings.TrimSuffix(filename, "modules/log/log.go")

	if projectPrefix == filename {
		// in case the source code file is moved, we can not trim the suffix, the code above should also be updated.
		panic("weblens logger unable to detect correct package prefix, please update file: " + filename)
	}

	osUrl := os.Getenv("OPENSEARCH_URL")
	osUser := os.Getenv("OPENSEARCH_USER")
	osPass := os.Getenv("OPENSEARCH_PASSWORD")

	var err error
	if osUrl != "" && osUser != "" && osPass != "" {
		OpenSearchClient, err = NewOpenSearchClient(osUrl, osUser, osPass)
	}

	if err != nil {
		panic(err)
	}

	zerolog.TimestampFieldName = "@timestamp"
	zerolog.ErrorStackMarshaler = MarshalStack
	zerolog.ErrorStackFieldName = "traceback"

	NewZeroLogger()
}

type ErrUnwrapHook struct{}

func (h ErrUnwrapHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	fmt.Println("ErrUnwrapHook", e, level, msg)
}

var logger zerolog.Logger = zerolog.Nop()

type LogOpts struct {
	NoOpenSearch bool
}

func compileLogOpts(o ...LogOpts) LogOpts {
	opts := LogOpts{}

	for _, opt := range o {
		if opt.NoOpenSearch {
			opts.NoOpenSearch = true
		}
	}

	return opts
}

func NopLogger() zerolog.Logger {
	nop := zerolog.Nop()

	return nop
}

func NewZeroLogger(opts ...LogOpts) *zerolog.Logger {
	o := compileLogOpts(opts...)

	if logger.GetLevel() != zerolog.Disabled && len(opts) == 0 {
		l := logger.With().Logger()

		return &l
	}

	config := config.GetConfig()

	var localLogger io.Writer
	if config.LogFormat == "dev" {
		localLogger = WLConsoleLogger{}
	} else {
		localLogger = os.Stdout
	}

	writers := []io.Writer{localLogger}

	if !o.NoOpenSearch && OpenSearchClient != nil {
		opnsIndex := os.Getenv("OPENSEARCH_INDEX")
		oLog := NewOpensearchLogger(OpenSearchClient, opnsIndex)
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
	log := zerolog.New(multi).Level(config.LogLevel).With().Timestamp().Caller().Str("weblens_build_version", wl_version).Logger()

	if len(opts) == 0 {
		zerolog.SetGlobalLevel(config.LogLevel)
		logger = log
		zlog.Logger = log
		log.Info().Msgf("Weblens logger initialized [%s][%s]", log.GetLevel(), config.LogFormat)
	}

	return &log
}

func GlobalLogger() *zerolog.Logger {
	l := logger
	if logger.GetLevel() == zerolog.Disabled {
		l = NopLogger()
	}

	return &l
}

type loggerContextKey struct{}

func WithContext(ctx context.Context, l *zerolog.Logger) context.Context {
	if ctx == nil {
		return ctx
	}

	ctx = context.WithValue(ctx, loggerContextKey{}, l)

	return ctx
}

func FromContext(ctx context.Context) *zerolog.Logger {
	l, ok := ctx.Value(loggerContextKey{}).(*zerolog.Logger)
	if !ok {
		return GlobalLogger()
	}

	return l
}

func FromContextOk(ctx context.Context) (*zerolog.Logger, bool) {
	l, ok := ctx.Value(loggerContextKey{}).(*zerolog.Logger)
	return l, ok
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

	// if level == zerolog.TraceLevel.String() || true {
	// 	// ignoredKeys := []string{"level", "error", "message", "weblens_build_version", "caller", "@timestamp", "traceback", "instance"}
	// 	allowedKeys := []string{"task_id"}
	// 	extras := ""
	//
	// 	for _, k := range allowedKeys {
	// 		v, ok := target[k]
	// 		if !ok {
	// 			continue
	// 		}
	//
	// 		extras += fmt.Sprintf("%s%s%s: %v ", BLUE, k, RESET, v)
	// 	}
	//
	// 	if extras != "" {
	// 		logMsg += "\n\t" + extras
	// 	}
	// }

	stackStr := ""
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

			line, _ := blockMap["line"].(string)

			source, okSource := blockMap["source"].(string)
			if okSource {
				if strings.HasPrefix(source, projectPrefix) {
					source = strings.TrimPrefix(source, projectPrefix)
				} else {
					source = ".../" + filepath.Base(source)
				}

				fileAndLine := source + ":" + line
				function += "()"
				stackStr += fmt.Sprintf("\t%s%-40s %s%30s\n", BLUE, fileAndLine, RESET, function)
			}
		}
	}

	levelColor := ""

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
