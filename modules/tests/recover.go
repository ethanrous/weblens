package tests

import (
	"context"
	"testing"

	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
)

func Setup(t *testing.T) context.Context {
	t.Helper()

	l := log.NewZeroLogger().With().Str("test_name", t.Name()).Logger()
	ctx := log.WithContext(t.Context(), &l)

	return ctx
}

func Recover(t *testing.T) {
	if rvr := recover(); rvr != nil {
		err, ok := rvr.(error)
		if !ok {
			err = errors.Errorf("Non-error panic in test: %v", rvr)
		}

		err = errors.WithStack(err)
		log.FromContext(t.Context()).Error().Stack().Err(err).Msg("Test failed:")

		t.Errorf("Test failed: %v", err)
	}
}
