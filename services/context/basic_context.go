package context

import (
	"context"

	"github.com/ethanrous/weblens/modules/log"
	"github.com/rs/zerolog"
)

type BasicContext struct {
	context.Context
}

func NewBasicContext(ctx context.Context, logger *zerolog.Logger) BasicContext {
	ctx = log.WithContext(ctx, logger)
	return BasicContext{
		Context: ctx,
	}
}

func (b BasicContext) Log() *zerolog.Logger {
	return log.FromContext(b.Context)
}

func (b BasicContext) WithValue(key, value any) BasicContext {
	b.Context = context.WithValue(b.Context, key, value)

	return b
}
