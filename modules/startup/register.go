// Package startup provides a mechanism for registering and running initialization functions during application startup.
package startup

import (
	"context"
	"reflect"
	"runtime"
	"time"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlog"
	"github.com/rs/zerolog"
)

// HookFunc is a function that performs initialization tasks during application startup.
type HookFunc func(context.Context, config.Provider) error

var startups []HookFunc

// RegisterHook adds a startup function to be executed during application initialization.
func RegisterHook(f HookFunc) {
	startups = append(startups, f)
}

// ErrDeferStartup signals that a startup function should be deferred and run later.
var ErrDeferStartup = wlerrors.New("defer startup")

// RunStartups executes all registered startup functions in order, supporting deferral.
func RunStartups(ctx context.Context, cnf config.Provider) error {
	wlog.FromContext(ctx).Info().Interface("startup_configuration", cnf).Msg("Running startup functions")

	start := time.Now()

	toRun := startups
	for len(toRun) != 0 {
		var startup HookFunc

		startup, toRun = toRun[0], toRun[1:]

		wlog.FromContext(ctx).Debug().Func(func(e *zerolog.Event) {
			funcName := runtime.FuncForPC(reflect.ValueOf(startup).Pointer()).Name()
			e.Interface("startup_function", funcName).Msg("Running startup function")
		})

		if err := startup(ctx, cnf); err != nil {
			if wlerrors.Is(err, ErrDeferStartup) {
				if len(toRun) == 0 {
					// Grab the function name for better error reporting
					funcName := runtime.FuncForPC(reflect.ValueOf(startup).Pointer()).Name()

					return wlerrors.Errorf("startup function [%s] requested to be deferred, but there are no more startups to run", funcName)
				}

				// Defer the startup
				toRun = append(toRun, startup)

				continue
			}

			return err
		}
	}

	wlog.FromContext(ctx).Info().Dur("startup_hooks_total_duration_ms", time.Since(start)).Msg("Completed startup functions")

	return nil
}
