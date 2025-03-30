package context

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

type BasicContext struct {
	context.Context
	Logger *zerolog.Logger
}

func (b *BasicContext) WithTimeout(d time.Duration) (*BasicContext, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(b.Context, d)

	return &BasicContext{
		Context: ctx,
		Logger:  b.Logger,
	}, cancel
}

func (b *BasicContext) Log() *zerolog.Logger {
	return b.Logger
}
