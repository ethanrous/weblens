package wlcontext

import (
	"context"

	"github.com/rs/zerolog"
)

// LoggerContext provides access to a zerolog logger instance.
type LoggerContext interface {
	context.Context
	Log() *zerolog.Logger
}
