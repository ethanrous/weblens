// Package context provides custom error types for context-related operations.
package context

import "fmt"

// CanceledError represents an error that occurred due to context cancellation.
type CanceledError struct {
	// Message provides additional context about the error.
	Message string
}

// Error implements the error interface for CanceledError.
func (e *CanceledError) Error() string {
	return fmt.Sprintf("context canceled: %s", e.Message)
}

// NewCanceledError creates a new CanceledError with the given message.
func NewCanceledError(message string) error {
	return &CanceledError{Message: message}
}
