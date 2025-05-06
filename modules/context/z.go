package context

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var ToZ func(c context.Context) ContextZ

type zkey struct{}

// ContextZ is a context that consolidates all other context interfaces.
type ContextZ interface {
	context.Context
	DatabaseContext
	DispatcherContext
	LoggerContext
	NotifierContext

	GetTowerId() string
	ClearCache()
	WithContext(ctx context.Context) context.Context
}

type AppContexter interface {
	AppCtx() ContextZ
}

var _ LoggerContext = &noplogger{}

type noplogger struct{ context.Context }

func (n *noplogger) Log() *zerolog.Logger {
	return &log.Logger
}

func (n *noplogger) WithLogger(zerolog.Logger) {}

func Background() LoggerContext {
	return &noplogger{context.Background()}
}
