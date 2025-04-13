package db

import (
	"fmt"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func WrapError(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf(format, a...)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return errors.WithStack(&NotFoundError{msg})
	} else {
		return fmt.Errorf("unknown database error: %s: %w", msg, err)
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
