package wlog

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

var projectPrefix string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	projectPrefix = strings.TrimSuffix(filename, "modules/wlog/log.go")

	if projectPrefix == filename {
		// in case the source code file is moved, we can not trim the suffix, the code above should also be updated.
		panic("weblens logger unable to detect correct package prefix, please update file: " + filename)
	}

	zerolog.TimestampFieldName = "@timestamp"
	zerolog.ErrorStackMarshaler = MarshalStack
	zerolog.ErrorStackFieldName = "traceback"
}

// ErrUnwrapHook is a zerolog hook for unwrapping errors.
type ErrUnwrapHook struct{}

// Run executes the hook on log events.
func (h ErrUnwrapHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	fmt.Println("ErrUnwrapHook", e, level, msg)
}

var logger zerolog.Logger = zerolog.Nop()
var loggerMu sync.RWMutex

// CreateOpts configures logging options.
type CreateOpts struct {
	Level   zerolog.Level
	LogFile string
}

func compileCreateLogOpts(o ...CreateOpts) CreateOpts {
	opts := CreateOpts{}
	opts.Level = zerolog.Disabled

	for _, opt := range o {
		if opt.Level != zerolog.Disabled {
			opts.Level = opt.Level
		}

		if opt.LogFile != "" {
			opts.LogFile = opt.LogFile
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

	loggerMu.RLock()

	if logger.GetLevel() != zerolog.Disabled && len(opts) == 0 {
		l := logger.With().Logger()

		loggerMu.RUnlock()

		return &l
	}

	loggerMu.RUnlock()

	outputLocation := os.Stdout

	if o.LogFile != "" {
		var err error

		outputLocation, err = os.OpenFile(o.LogFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(wlerrors.Errorf("failed to open log file %s: %w", o.LogFile, err))
		}
	}

	config := config.GetConfig()

	var logWriter io.Writer
	if config.LogFormat == "dev" {
		logWriter = newDevLogger(outputLocation)
	} else {
		logWriter = outputLocation
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

	level := zerolog.InfoLevel
	if o.Level != zerolog.Disabled {
		level = o.Level
	}

	// Create a new logger instance with the specified output, log level, and build version context
	log := zerolog.New(logWriter).Level(level).With().Timestamp().Caller().Str("weblens_build_version", wlVersion).Logger()

	// If no options are provided, set the global loggers
	if len(opts) == 0 {
		zerolog.SetGlobalLevel(config.LogLevel)

		loggerMu.Lock()

		// Set as our "global" logger
		logger = log

		// Set as the zerolog global logger
		zlog.Logger = log

		loggerMu.Unlock()

		log.Info().Msgf("Weblens logger initialized [%s][%s]", log.GetLevel(), config.LogFormat)
	} else {
		log.Debug().Msgf("Created new Weblens logger [%s][%s]", log.GetLevel(), config.LogFormat)
	}

	return &log
}

// SetLogLevel sets the global log level.
func SetLogLevel(level zerolog.Level) {
	config.SetLogLevel(level)
	zerolog.SetGlobalLevel(level)

	loggerMu.Lock()

	logger = logger.Level(level)

	loggerMu.Unlock()
}

// GlobalLogger returns a copy of the global logger instance.
// The returned logger is safe to use concurrently and can have
// UpdateContext called on it without racing with other goroutines.
func GlobalLogger() *zerolog.Logger {
	loggerMu.RLock()

	if logger.GetLevel() == zerolog.Disabled {
		loggerMu.RUnlock()

		NewZeroLogger()

		loggerMu.RLock()
	}

	defer loggerMu.RUnlock()

	// Use With().Logger() to create a copy
	// that doesn't share the context buffer
	l := logger.With().Logger()

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
