package context

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

type BasicContext struct {
	context.Context
	Log *zerolog.Logger
}

func (b BasicContext) WithTimeout(d time.Duration) (BasicContext, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(b.Context, d)

	return BasicContext{
		Context: ctx,
		Log:     b.Log,
	}, cancel
}
