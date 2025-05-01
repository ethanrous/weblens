package context

import (
	"context"

	"github.com/rs/zerolog"
)

type BasicContext struct {
	context.Context
	Logger zerolog.Logger
}

func NewBasicContext(ctx context.Context, logger zerolog.Logger) BasicContext {
	return BasicContext{
		Context: ctx,
		Logger:  logger,
	}
}

func (b BasicContext) Log() *zerolog.Logger {
	return &b.Logger
}

func (b BasicContext) WithValue(key, value any) BasicContext {
	b.Context = context.WithValue(b.Context, key, value)

	return b
}

func (b BasicContext) WithLogger(l zerolog.Logger) {
	b.Logger = l
}
