package db

import (
	"fmt"

	"github.com/ethanrous/weblens/modules/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func WrapError(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf(format, a...)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return errors.WithStack(&NotFoundError{msg})
	} else if errors.Is(err, mongo.ErrNilCursor) {
		return errors.WithStack(&AlreadyExistsError{msg})
	} else {
		return errors.WithStack(fmt.Errorf("unknown database error: %s: %w", msg, err))
	}
}

// NotFoundError represents a generic "not found" error in the database.
type NotFoundError struct {
	// Message provides additional context about the error.
	Message string
}

// Error implements the error interface for NotFoundError.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s: not found", e.Message)
}

// NewNotFoundError creates a new NotFoundError with the given message.
func NewNotFoundError(message string) error {
	return &NotFoundError{Message: message}
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	var notFoundErr *NotFoundError
	if errors.As(err, &notFoundErr) {
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

func IsAlreadyExists(err error) bool {
	var alreadyExistsErr *AlreadyExistsError
	if errors.As(err, &alreadyExistsErr) {
		return true
	}
	return false
}
