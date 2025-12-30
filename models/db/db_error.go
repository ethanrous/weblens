package db

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"go.mongodb.org/mongo-driver/mongo"
)

// WrapError wraps a database error with additional context and converts it to a domain-specific error type.
func WrapError(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf(format, a...)

	if wlerrors.Is(err, mongo.ErrNoDocuments) {
		return wlerrors.WithStack(&NotFoundError{msg})
	} else if strings.Contains(err.Error(), "duplicate key error") {
		return wlerrors.WithStack(&AlreadyExistsError{msg})
	} else if strings.HasSuffix(err.Error(), "context canceled") {
		return wlerrors.WithStack(wlcontext.NewCanceledError(msg))
	}

	return wlerrors.WithStack(fmt.Errorf("unknown database error: %s: %w", msg, err))
}

var _ error = &NotFoundError{}

// NotFoundError represents a generic "not found" error in the database.
type NotFoundError struct {
	// Message provides additional context about the error.
	Message string
}

// Error implements the error interface for NotFoundError.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s: not found", e.Message)
}

// Status returns the HTTP status code for a NotFoundError.
func (e *NotFoundError) Status() int {
	return http.StatusNotFound
}

// NewNotFoundError creates a new NotFoundError with the given message.
func NewNotFoundError(message string) error {
	return &NotFoundError{Message: message}
}

// IsNotFound checks if an error is a NotFoundError.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	var notFoundErr *NotFoundError
	if wlerrors.As(err, &notFoundErr) {
		return true
	}

	return false
}

// AlreadyExistsError represents a generic "already exists" conflict error in the database.
type AlreadyExistsError struct {
	// Message provides additional context about the error.
	Message string
}

// Error implements the error interface for AlreadyExistsError.
func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("already exists: %s", e.Message)
}

// NewAlreadyExistsError creates a new AlreadyExistsError with the given message.
func NewAlreadyExistsError(message string) error {
	return &AlreadyExistsError{Message: message}
}

// IsAlreadyExists checks if an error is an AlreadyExistsError.
func IsAlreadyExists(err error) bool {
	var alreadyExistsErr *AlreadyExistsError
	if wlerrors.As(err, &alreadyExistsErr) {
		return true
	}

	return false
}
