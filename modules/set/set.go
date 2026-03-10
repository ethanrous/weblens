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

// Add adds an item(s) to the Set.
func (s Set[T]) Add(items ...T) {
	for _, item := range items {
		s[item] = struct{}{}
	}
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

// Len returns the number of items in the Set.
func (s Set[T]) Len() int {
	return len(s)
}

// Union returns a new Set containing all unique items from both sets.
func (s Set[T]) Union(other Set[T]) Set[T] {
	result := make(Set[T], s.Len()+other.Len())

	for item := range s {
		result.Add(item)
	}

	for item := range other {
		result.Add(item)
	}

	return result
}

// Intersection returns a new Set containing only items that are present in both sets.
func (s Set[T]) Intersection(other Set[T]) Set[T] {
	result := New[T]()

	for item := range s {
		if other.Has(item) {
			result.Add(item)
		}
	}

	return result
}

// SymmetricDifference returns a new Set containing items that are present in either set but not in both.
func SymmetricDifference[T comparable](s1, s2 Set[T]) Set[T] {
	result := New[T]()

	for item := range s1 {
		if !s2.Has(item) {
			result.Add(item)
		}
	}

	for item := range s2 {
		if !s1.Has(item) {
			result.Add(item)
		}
	}

	return result
}
