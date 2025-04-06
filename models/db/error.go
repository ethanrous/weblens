package db

import "fmt"

// NotFoundError represents a generic "not found" error in the database.
type NotFoundError struct {
	// Message provides additional context about the error.
	Message string
}

// Error implements the error interface for NotFoundError.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %s", e.Message)
}

// NewNotFoundError creates a new NotFoundError with the given message.
func NewNotFoundError(message string) error {
	return &NotFoundError{Message: message}
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
