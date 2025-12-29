package context

import (
	"context"

	"github.com/ethanrous/weblens/modules/log"
	"github.com/rs/zerolog"
)

// BasicContext wraps a standard context.Context with additional logging capabilities.
type BasicContext struct {
	context.Context
}

// NewBasicContext creates a new BasicContext with the provided logger attached.
func NewBasicContext(ctx context.Context, logger *zerolog.Logger) BasicContext {
	ctx = log.WithContext(ctx, logger)

	return BasicContext{
		Context: ctx,
	}
}

// Log retrieves the logger instance from the BasicContext.
func (b BasicContext) Log() *zerolog.Logger {
	return log.FromContext(b.Context)
}

// WithValue returns a copy of BasicContext with the specified key-value pair added.
func (b BasicContext) WithValue(key, value any) BasicContext {
	b.Context = context.WithValue(b.Context, key, value)

	return b
}
