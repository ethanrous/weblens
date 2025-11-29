package context

import "fmt"

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
