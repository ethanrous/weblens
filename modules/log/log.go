package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

// OpenSearchClient is the global OpenSearch client for logging.
var OpenSearchClient *opensearch.Client

var projectPrefix string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	projectPrefix = strings.TrimSuffix(filename, "modules/log/log.go")

	if projectPrefix == filename {
		// in case the source code file is moved, we can not trim the suffix, the code above should also be updated.
		panic("weblens logger unable to detect correct package prefix, please update file: " + filename)
	}

	osURL := os.Getenv("OPENSEARCH_URL")
	osUser := os.Getenv("OPENSEARCH_USER")
	osPass := os.Getenv("OPENSEARCH_PASSWORD")

	var err error
	if osURL != "" && osUser != "" && osPass != "" {
		OpenSearchClient, err = NewOpenSearchClient(osURL, osUser, osPass)
	}

	if err != nil {
		fmt.Println("Error initializing OpenSearch client for logging:", err)
	}

	zerolog.TimestampFieldName = "@timestamp"
	zerolog.ErrorStackMarshaler = MarshalStack
	zerolog.ErrorStackFieldName = "traceback"

	NewZeroLogger()
}

// ErrUnwrapHook is a zerolog hook for unwrapping errors.
type ErrUnwrapHook struct{}

// Run executes the hook on log events.
func (h ErrUnwrapHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	fmt.Println("ErrUnwrapHook", e, level, msg)
}

var logger zerolog.Logger = zerolog.Nop()

// CreateOpts configures logging options.
type CreateOpts struct {
	NoOpenSearch bool
}

func compileCreateLogOpts(o ...CreateOpts) CreateOpts {
	opts := CreateOpts{}

	for _, opt := range o {
		if opt.NoOpenSearch {
			opts.NoOpenSearch = true
		}
	}

	return opts
}

// NopLogger returns a no-op logger that discards all log messages.
func NopLogger() zerolog.Logger {
	nop := zerolog.Nop()

	return nop
}

// NewZeroLogger creates a new zerolog logger with the given options.
func NewZeroLogger(opts ...CreateOpts) *zerolog.Logger {
	o := compileCreateLogOpts(opts...)

	if logger.GetLevel() != zerolog.Disabled && len(opts) == 0 {
		l := logger.With().Logger()

		return &l
	}

	config := config.GetConfig()

	var localLogger io.Writer
	if config.LogFormat == "dev" {
		localLogger = newDevLogger()
	} else {
		localLogger = os.Stdout
	}

	writers := []io.Writer{localLogger}

	if !o.NoOpenSearch && OpenSearchClient != nil {
		opnsIndex := os.Getenv("OPENSEARCH_INDEX")
		oLog := NewOpensearchLogger(OpenSearchClient, opnsIndex)
		writers = append(writers, oLog)
	}

	wlVersion := os.Getenv("WEBLENS_BUILD_VERSION")
	if wlVersion == "" {
		wlVersion = "unknown"

		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			for _, v := range buildInfo.Settings {
				if v.Key == "vcs.revision" {
					wlVersion = v.Value

					break
				}
			}
		}
	}

	multi := zerolog.MultiLevelWriter(writers...)
	log := zerolog.New(multi).Level(config.LogLevel).With().Timestamp().Caller().Str("weblens_build_version", wlVersion).Logger()

	if len(opts) == 0 {
		zerolog.SetGlobalLevel(config.LogLevel)

		logger = log
		zlog.Logger = log
		log.Info().Msgf("Weblens logger initialized [%s][%s]", log.GetLevel(), config.LogFormat)
	}

	return &log
}

// GlobalLogger returns the global logger instance.
func GlobalLogger() *zerolog.Logger {
	l := logger
	if logger.GetLevel() == zerolog.Disabled {
		l = NopLogger()
	}

	return &l
}

type loggerContextKey struct{}

// WithContext adds a logger to the context.
func WithContext(ctx context.Context, l *zerolog.Logger) context.Context {
	if ctx == nil {
		return ctx
	}

	ctx = context.WithValue(ctx, loggerContextKey{}, l)

	return ctx
}

// FromContext extracts a logger from the context, or returns the global logger.
func FromContext(ctx context.Context) *zerolog.Logger {
	l, ok := ctx.Value(loggerContextKey{}).(*zerolog.Logger)
	if !ok {
		return GlobalLogger()
	}

	return l
}

// FromContextOk extracts a logger from the context and returns whether it was present.
func FromContextOk(ctx context.Context) (*zerolog.Logger, bool) {
	l, ok := ctx.Value(loggerContextKey{}).(*zerolog.Logger)

	return l, ok
}

// ShowStackTrace is a placeholder for stack trace visualization.
func ShowStackTrace() {

}
