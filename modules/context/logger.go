package context

import (
	"context"

	"github.com/rs/zerolog"
)

type LoggerContext interface {
	context.Context
	Log() *zerolog.Logger
}
