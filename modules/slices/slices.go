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

func OnlyUnique[T comparable](s []T) (rs []T) {
	tmpMap := make(map[T]struct{}, len(s))
	for _, t := range s {
		tmpMap[t] = struct{}{}
	}
	return slices.Collect(maps.Keys(tmpMap))
}

func AddToSet[T comparable](set []T, add ...T) []T {
	for _, a := range add {
		if !slices.Contains(set, a) {
			set = append(set, a)
		}
	}
	return set
}

func InsertFunc[S ~[]T, T any](ts S, t T, cmp func(a T, b T) int) S {
	i, _ := slices.BinarySearchFunc(ts, t, cmp) // find slot
	return slices.Insert(ts, i, t)
}

func Contains[S ~[]T, T comparable](ts S, t T) bool {
	return slices.Contains(ts, t) // find slot

}

func ContainsS[S ~[]T, T cmp.Ordered](ts S, t T) bool {
	i, _ := slices.BinarySearch(ts, t) // find slot
	return i >= 0
}

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

func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

func Filter[S ~[]T, T any](ts S, fn func(t T) bool) []T {
	var result []T
	for _, t := range ts {
		if fn(t) {
			result = append(result, t)
		}
	}
	return result
}

func FilterMap[T, V any](ts []T, fn func(T) (V, bool)) []V {
	var result []V = make([]V, 0)
	for _, t := range ts {
		res, y := fn(t)
		if y {
			result = append(result, res)
		}
	}
	return result
}

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
