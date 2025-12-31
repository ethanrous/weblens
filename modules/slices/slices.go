// Package slices provides utility functions for working with slices in Go.
package slices

import (
	"cmp"
	"maps"
	"slices"
)

// Yoink See Banish. Yoink is the same as Banish, but returns the value at i
// in addition to the shortened slice.
func Yoink[T any](s []T, i int) ([]T, T) {
	t := s[i]
	n := append(s[:i], s[i+1:]...)

	return n, t
}

// OnlyUnique returns a slice containing only the unique elements from s.
func OnlyUnique[T comparable](s []T) (rs []T) {
	tmpMap := make(map[T]struct{}, len(s))
	for _, t := range s {
		tmpMap[t] = struct{}{}
	}

	return slices.Collect(maps.Keys(tmpMap))
}

// AddToSet adds elements to a set (slice with unique values) if they don't already exist.
func AddToSet[T comparable](set []T, add ...T) []T {
	for _, a := range add {
		if !slices.Contains(set, a) {
			set = append(set, a)
		}
	}

	return set
}

// InsertFunc inserts an element into a sorted slice using a comparison function.
func InsertFunc[S ~[]T, T any](ts S, t T, cmp func(a T, b T) int) S {
	i, _ := slices.BinarySearchFunc(ts, t, cmp) // find slot

	return slices.Insert(ts, i, t)
}

// Contains returns true if the slice contains the given element.
func Contains[S ~[]T, T comparable](ts S, t T) bool {
	return slices.Contains(ts, t) // find slot
}

// ContainsS returns true if the sorted slice contains the given element using binary search.
func ContainsS[S ~[]T, T cmp.Ordered](ts S, t T) bool {
	i, _ := slices.BinarySearch(ts, t) // find slot

	return i >= 0
}

// Diff returns elements in s1 that are not in s2.
func Diff[T comparable](s1 []T, s2 []T) []T {
	if len(s1) < len(s2) {
		s1, s2 = s2, s1
	}

	var res []T

	for _, t := range s1 {
		if !slices.Contains(s2, t) {
			res = append(res, t)
		}
	}

	return res
}

// Map transforms each element of a slice using the provided function.
func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}

	return result
}

// MapI transforms each element of a slice with access to the index.
func MapI[T, V any](ts []T, fn func(T, int) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t, i)
	}

	return result
}

// SortFunc sorts a slice in place using a comparison function and returns it.
func SortFunc[S ~[]T, T any](ts S, cmp func(a T, b T) int) S {
	slices.SortFunc(ts, cmp)

	return ts
}

// BinarySearchFunc searches for an element in a sorted slice using a comparison function.
func BinarySearchFunc[S ~[]E, E, T any](x S, target T, cmp func(E, T) int) (int, bool) {
	return slices.BinarySearchFunc(x, target, cmp)
}

// BinarySearchIndexFunc returns the insertion index for target in a sorted slice.
func BinarySearchIndexFunc[S ~[]E, E, T any](x S, target T, cmp func(E, T) int) int {
	n := len(x)

	i, j := 0, n
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		// i ≤ h < j
		if cmp(x[h], target) < 0 {
			i = h + 1 // preserves x[i-1] < target
		} else {
			j = h // preserves x[j] >= target
		}
	}

	return i
}

// Filter returns elements that satisfy the predicate function.
func Filter[S ~[]T, T any](ts S, fn func(t T) bool) []T {
	var result []T

	for _, t := range ts {
		if fn(t) {
			result = append(result, t)
		}
	}

	return result
}

// FilterMap transforms and filters elements in a single pass.
func FilterMap[T, V any](ts []T, fn func(T) (V, bool)) []V {
	result := make([]V, 0)

	for _, t := range ts {
		res, y := fn(t)
		if y {
			result = append(result, res)
		}
	}

	return result
}

// Reduce applies a function to each element, accumulating a single result.
func Reduce[T, A any](ts []T, fn func(T, A) A, acc A) A {
	for _, t := range ts {
		acc = fn(t, acc)
	}

	return acc
}

// Convert Perform type assertion on slice
func Convert[V, T any](ts []T) []V {
	vs := make([]V, len(ts))
	if len(ts) == 0 {
		return vs
	}

	for i := range ts {
		vs[i] = any(ts[i]).(V)
	}

	return vs
}

// Index returns the index of the first occurrence of v in s, or -1 if not present.
func Index[S ~[]E, E comparable](s S, v E) int {
	return slices.Index(s, v)
}

// IndexFunc returns the index of the first element satisfying f(e), or -1 if none do.
func IndexFunc[S ~[]E, E any](s S, f func(E) bool) int {
	return slices.IndexFunc(s, f)
}
