package context

import (
	"context"
	"sync"

	"github.com/ethanrous/weblens/modules/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	if ToZ == nil {
		ToZ = func(c context.Context) Z {
			if c == nil {
				return nil
			}

			if ctx, ok := c.(Z); ok {
				return ctx
			}

			panic("ToZ: context is not a ContextZ")
		}
	}
}

// ToZ converts a context.Context to a ContextZ, panicking if the conversion fails.
var ToZ func(c context.Context) Z

// Z is a context that consolidates all other context interfaces.
type Z interface {
	context.Context
	DatabaseContext
	LoggerContext

	WithContext(ctx context.Context) context.Context
}

// AppContexter provides access to the application's ContextZ.
type AppContexter interface {
	AppCtx() Z
}

var _ LoggerContext = &noplogger{}

type noplogger struct{ context.Context }

func (n *noplogger) Log() *zerolog.Logger {
	return &log.Logger
}

func (n *noplogger) WithLogger(zerolog.Logger) {}

// Background returns a LoggerContext with a no-op logger for use as a background context.
func Background() LoggerContext {
	return &noplogger{context.Background()}
}

// RequestDoerKey is the context key used to store the user making a request.
const RequestDoerKey = "doer"

// Doer provides the identity of the user performing an action.
type Doer interface {
	Doer() string
}

// WgKey is the context key used to store a WaitGroup for coordinating goroutines.
const WgKey = "wg"

// AddToWg increments the WaitGroup stored in the context by one.
func AddToWg(ctx context.Context) error {
	if ctx == nil {
		return errors.New("AddToWg: context is nil")
	}

	wg, ok := ctx.Value(WgKey).(*sync.WaitGroup)
	if !ok {
		return errors.New("AddToWg: context does not contain a WaitGroup")
	}

	wg.Add(1)

	return nil
}

// WgDone decrements the WaitGroup stored in the context by one.
func WgDone(ctx context.Context) error {
	if ctx == nil {
		return errors.New("AddToWg: context is nil")
	}

	wg, ok := ctx.Value(WgKey).(*sync.WaitGroup)
	if !ok {
		return errors.New("AddToWg: context does not contain a WaitGroup")
	}

	wg.Done()

	return nil
}
