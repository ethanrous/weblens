// Package set provides a generic set data structure for storing unique values.
package set

// Set is a generic set data structure for storing unique comparable values.
type Set[T comparable] map[T]struct{}

// New creates a new Set containing the provided items.
func New[T comparable](items ...T) Set[T] {
	s := make(Set[T], len(items))
	for _, item := range items {
		s.Add(item)
	}

	return s
}

// Add adds an item to the Set.
func (s Set[T]) Add(item T) {
	s[item] = struct{}{}
}

// Remove removes an item from the Set.
func (s Set[T]) Remove(item T) {
	delete(s, item)
}

// Has returns true if the Set contains the specified item.
func (s Set[T]) Has(item T) bool {
	_, exists := s[item]

	return exists
}
