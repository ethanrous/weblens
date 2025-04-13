package context

import "context"

// ContextZ is a context that consolidates all other context interfaces.
type ContextZ interface {
	context.Context
	DatabaseContext
	DispatcherContext
	LoggerContext
	NotifierContext
}

type AppContexter interface {
	AppCtx() ContextZ
}
