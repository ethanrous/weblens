//go:build test

package wlerrors_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("creates error with message", func(t *testing.T) {
		err := wlerrors.New("test error")
		assert.NotNil(t, err)
		assert.Equal(t, "test error", err.Error())
	})

	t.Run("creates error with empty message", func(t *testing.T) {
		err := wlerrors.New("")
		assert.NotNil(t, err)
		assert.Equal(t, "", err.Error())
	})
}

func TestErrorf(t *testing.T) {
	t.Run("creates formatted error", func(t *testing.T) {
		err := wlerrors.Errorf("error: %s %d", "test", 42)
		assert.NotNil(t, err)
		assert.Equal(t, "error: test 42", err.Error())
	})

	t.Run("creates error without format args", func(t *testing.T) {
		err := wlerrors.Errorf("simple error")
		assert.NotNil(t, err)
		assert.Equal(t, "simple error", err.Error())
	})
}

func TestWithStack(t *testing.T) {
	t.Run("adds stack to error", func(t *testing.T) {
		original := errors.New("original error")
		wrapped := wlerrors.WithStack(original)
		assert.NotNil(t, wrapped)
		assert.Equal(t, "original error", wrapped.Error())
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		wrapped := wlerrors.WithStack(nil)
		assert.Nil(t, wrapped)
	})

	t.Run("preserves original error through Unwrap", func(t *testing.T) {
		original := errors.New("original")
		wrapped := wlerrors.WithStack(original)
		assert.True(t, errors.Is(wrapped, original))
	})
}

func TestWrap(t *testing.T) {
	t.Run("wraps error with message", func(t *testing.T) {
		original := errors.New("original error")
		wrapped := wlerrors.Wrap(original, "wrapper message")
		assert.NotNil(t, wrapped)
		assert.Contains(t, wrapped.Error(), "wrapper message")
		assert.Contains(t, wrapped.Error(), "original error")
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		wrapped := wlerrors.Wrap(nil, "message")
		assert.Nil(t, wrapped)
	})

	t.Run("preserves original error through chain", func(t *testing.T) {
		original := errors.New("original")
		wrapped := wlerrors.Wrap(original, "wrapped")
		assert.True(t, errors.Is(wrapped, original))
	})
}

func TestWrapf(t *testing.T) {
	t.Run("wraps error with formatted message", func(t *testing.T) {
		original := errors.New("original")
		wrapped := wlerrors.Wrapf(original, "context: %s", "details")
		assert.NotNil(t, wrapped)
		assert.Contains(t, wrapped.Error(), "context: details")
		assert.Contains(t, wrapped.Error(), "original")
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		wrapped := wlerrors.Wrapf(nil, "format %s", "arg")
		assert.Nil(t, wrapped)
	})
}

func TestWithMessage(t *testing.T) {
	t.Run("adds message to error", func(t *testing.T) {
		original := errors.New("original")
		wrapped := wlerrors.WithMessage(original, "additional context")
		assert.NotNil(t, wrapped)
		assert.Contains(t, wrapped.Error(), "additional context")
		assert.Contains(t, wrapped.Error(), "original")
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		wrapped := wlerrors.WithMessage(nil, "message")
		assert.Nil(t, wrapped)
	})
}

func TestWithMessagef(t *testing.T) {
	t.Run("adds formatted message to error", func(t *testing.T) {
		original := errors.New("original")
		wrapped := wlerrors.WithMessagef(original, "context: %d", 42)
		assert.NotNil(t, wrapped)
		assert.Contains(t, wrapped.Error(), "context: 42")
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		wrapped := wlerrors.WithMessagef(nil, "format %s", "arg")
		assert.Nil(t, wrapped)
	})
}

func TestCause(t *testing.T) {
	t.Run("returns root cause", func(t *testing.T) {
		original := wlerrors.New("original")
		wrapped := wlerrors.Wrap(original, "layer1")
		doubleWrapped := wlerrors.Wrap(wrapped, "layer2")

		cause := wlerrors.Cause(doubleWrapped)
		assert.Equal(t, "original", cause.Error())
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		cause := wlerrors.Cause(nil)
		assert.Nil(t, cause)
	})

	t.Run("returns error itself if no cause", func(t *testing.T) {
		original := errors.New("no cause")
		cause := wlerrors.Cause(original)
		assert.Equal(t, original, cause)
	})
}

func TestStatusf(t *testing.T) {
	t.Run("creates status error", func(t *testing.T) {
		err := wlerrors.Statusf(http.StatusNotFound, "resource %s not found", "user")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "resource user not found")
	})

	t.Run("status can be extracted", func(t *testing.T) {
		err := wlerrors.Statusf(http.StatusBadRequest, "bad request")
		status, msg := wlerrors.AsStatus(err, 0)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Contains(t, msg, "bad request")
	})
}

func TestWrapStatus(t *testing.T) {
	t.Run("wraps error with status", func(t *testing.T) {
		original := errors.New("original error")
		wrapped := wlerrors.WrapStatus(http.StatusInternalServerError, original)
		assert.NotNil(t, wrapped)
		assert.Equal(t, "original error", wrapped.Error())
	})

	t.Run("status can be extracted", func(t *testing.T) {
		original := errors.New("original")
		wrapped := wlerrors.WrapStatus(http.StatusForbidden, original)
		status, _ := wlerrors.AsStatus(wrapped, 0)
		assert.Equal(t, http.StatusForbidden, status)
	})

	t.Run("preserves original error", func(t *testing.T) {
		original := errors.New("original")
		wrapped := wlerrors.WrapStatus(http.StatusNotFound, original)
		assert.True(t, errors.Is(wrapped, original))
	})
}

func TestAsStatus(t *testing.T) {
	t.Run("extracts status from status error", func(t *testing.T) {
		err := wlerrors.Statusf(http.StatusUnauthorized, "unauthorized")
		status, msg := wlerrors.AsStatus(err, http.StatusInternalServerError)
		assert.Equal(t, http.StatusUnauthorized, status)
		assert.Contains(t, msg, "unauthorized")
	})

	t.Run("returns default for non-status error", func(t *testing.T) {
		err := errors.New("regular error")
		status, msg := wlerrors.AsStatus(err, http.StatusInternalServerError)
		assert.Equal(t, http.StatusInternalServerError, status)
		assert.Equal(t, "regular error", msg)
	})

	t.Run("returns zero for nil error", func(t *testing.T) {
		status, msg := wlerrors.AsStatus(nil, http.StatusInternalServerError)
		assert.Equal(t, 0, status)
		assert.Equal(t, "", msg)
	})
}

func TestIs(t *testing.T) {
	t.Run("matches same error", func(t *testing.T) {
		err := wlerrors.New("error")
		assert.True(t, wlerrors.Is(err, err))
	})

	t.Run("matches wrapped error", func(t *testing.T) {
		original := wlerrors.New("original")
		wrapped := wlerrors.Wrap(original, "wrapped")
		assert.True(t, wlerrors.Is(wrapped, original))
	})

	t.Run("does not match different errors", func(t *testing.T) {
		err1 := wlerrors.New("error1")
		err2 := wlerrors.New("error2")
		assert.False(t, wlerrors.Is(err1, err2))
	})
}

func TestAs(t *testing.T) {
	t.Run("extracts status error", func(t *testing.T) {
		err := wlerrors.Statusf(http.StatusNotFound, "not found")
		var statusErr wlerrors.StatusErr
		found := wlerrors.As(err, &statusErr)
		assert.True(t, found)
		assert.Equal(t, http.StatusNotFound, statusErr.Status())
	})

	t.Run("returns false for non-matching type", func(t *testing.T) {
		err := errors.New("regular error")
		var statusErr wlerrors.StatusErr
		found := wlerrors.As(err, &statusErr)
		assert.False(t, found)
	})
}

func TestIgnore(t *testing.T) {
	t.Run("returns nil for ignored error", func(t *testing.T) {
		errToIgnore := wlerrors.New("ignore me")
		result := wlerrors.Ignore(errToIgnore, errToIgnore)
		assert.Nil(t, result)
	})

	t.Run("returns error if not in ignore list", func(t *testing.T) {
		err := wlerrors.New("don't ignore")
		other := wlerrors.New("other")
		result := wlerrors.Ignore(err, other)
		assert.Equal(t, err, result)
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		other := wlerrors.New("other")
		result := wlerrors.Ignore(nil, other)
		assert.Nil(t, result)
	})

	t.Run("ignores wrapped error", func(t *testing.T) {
		original := wlerrors.New("original")
		wrapped := wlerrors.Wrap(original, "wrapped")
		result := wlerrors.Ignore(wrapped, original)
		assert.Nil(t, result)
	})

	t.Run("ignores any of multiple errors", func(t *testing.T) {
		err1 := wlerrors.New("err1")
		err2 := wlerrors.New("err2")
		err3 := wlerrors.New("err3")

		result := wlerrors.Ignore(err2, err1, err2, err3)
		assert.Nil(t, result)
	})
}

func TestFrameFormat(t *testing.T) {
	t.Run("error can be formatted with stack", func(t *testing.T) {
		err := wlerrors.New("test error")
		formatted := fmt.Sprintf("%+v", err)
		assert.Contains(t, formatted, "test error")
		// Should contain stack trace info
		assert.Contains(t, formatted, "wlerrors_test")
	})
}

func TestStackTrace(t *testing.T) {
	t.Run("Errorf includes stack trace", func(t *testing.T) {
		err := wlerrors.Errorf("formatted error")
		formatted := fmt.Sprintf("%+v", err)
		assert.Contains(t, formatted, "formatted error")
		assert.Contains(t, formatted, ".go:")
	})
}
