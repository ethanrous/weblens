// Package wlatomic provides a generic atomic value wrapper that allows for thread-safe access to a value of any type.
package wlatomic

import "sync"

// AtomicValue is a generic atomic value wrapper that provides thread-safe access to a value of any type.
type AtomicValue[T any] struct {
	value T
	mu    sync.RWMutex
}

// New creates a new AtomicValue with the given initial value.
func New[T any](initial T) *AtomicValue[T] {
	return &AtomicValue[T]{value: initial}
}

// Load returns the current value in a thread-safe manner.
func (av *AtomicValue[T]) Load() T {
	av.mu.RLock()
	defer av.mu.RUnlock()

	return av.value
}

// Set updates the value in a thread-safe manner.
func (av *AtomicValue[T]) Set(newValue T) {
	av.mu.Lock()
	defer av.mu.Unlock()

	av.value = newValue
}
