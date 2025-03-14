package log

import (
	"io"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Deprecated: Use json logger instead.
var Error *log.Logger

var OpenSearchClient *opensearch.Client

func init() {

	// Error writes logs in the color red with "ERROR: " as prefix
	Error = log.New(os.Stdout, "\u001b[31m[ERROR] \u001b[0m", log.LstdFlags|log.Lshortfile)

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
}

// Run(e *Event, level Level, message string)

func NewZeroLogger() *zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var localLogger io.Writer
	if os.Getenv("LOG_FORMAT") == "dev" {
		localLogger = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	} else {
		localLogger = os.Stdout
	}

	loggers := []io.Writer{localLogger}

	osIndex := os.Getenv("OPENSEARCH_INDEX")
	if OpenSearchClient != nil {
		oLog := NewOpensearchLogger(OpenSearchClient, osIndex)
		loggers = append(loggers, oLog)
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

	multi := zerolog.MultiLevelWriter(loggers...)
	log := zerolog.New(multi).With().Timestamp().Caller().Str("weblens_build_version", wl_version).Logger()

	return &log
}
