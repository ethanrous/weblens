// Package tests provides utility functions for test setup and error recovery in test cases.
package tests

import (
	"context"
	"testing"

	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlog"
)

// Setup initializes a test context with logging configured for the test.
func Setup(t *testing.T) context.Context {
	t.Helper()

	l := wlog.NewZeroLogger().With().Str("test_name", t.Name()).Logger()
	ctx := wlog.WithContext(t.Context(), &l)

	return ctx
}

// Recover handles panic recovery in tests by converting panics to test errors with stack traces.
func Recover(t *testing.T) {
	if rvr := recover(); rvr != nil {
		err, ok := rvr.(error)
		if !ok {
			err = wlerrors.Errorf("Non-error panic in test: %v", rvr)
		}

		err = wlerrors.WithStack(err)
		wlog.FromContext(t.Context()).Error().Stack().Err(err).Msg("Test failed:")

		t.Errorf("Test failed: %v", err)
	}
}
