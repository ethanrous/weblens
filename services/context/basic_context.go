package context

import (
	"context"

	"github.com/rs/zerolog"
)

type BasicContext struct {
	context.Context
	Logger *zerolog.Logger
}

func NewBasicContext(ctx context.Context, logger *zerolog.Logger) *BasicContext {
	return &BasicContext{
		Context: ctx,
		Logger:  logger,
	}
}

func (b *BasicContext) Log() *zerolog.Logger {
	return b.Logger
}
