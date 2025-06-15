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
		ToZ = func(c context.Context) ContextZ {
			if c == nil {
				return nil
			}

			if ctx, ok := c.(ContextZ); ok {
				return ctx
			}

			panic("ToZ: context is not a ContextZ")
		}
	}
}

var ToZ func(c context.Context) ContextZ

type zkey struct{}

// ContextZ is a context that consolidates all other context interfaces.
type ContextZ interface {
	context.Context
	DatabaseContext
	LoggerContext

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

const RequestDoerKey = "doer"

type Doer interface {
	Doer() string
}

const WgKey = "wg"

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
