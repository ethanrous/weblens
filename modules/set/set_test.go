//go:build test

package set_test

import (
	"testing"

	"github.com/ethanrous/weblens/modules/set"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("creates empty set", func(t *testing.T) {
		s := set.New[string]()
		assert.Equal(t, 0, len(s))
	})

	t.Run("creates set with single item", func(t *testing.T) {
		s := set.New("item")
		assert.Equal(t, 1, len(s))
		assert.True(t, s.Has("item"))
	})

	t.Run("creates set with multiple items", func(t *testing.T) {
		s := set.New("a", "b", "c")
		assert.Equal(t, 3, len(s))
		assert.True(t, s.Has("a"))
		assert.True(t, s.Has("b"))
		assert.True(t, s.Has("c"))
	})

	t.Run("creates set with duplicate items", func(t *testing.T) {
		s := set.New("a", "b", "a", "c", "b")
		assert.Equal(t, 3, len(s))
		assert.True(t, s.Has("a"))
		assert.True(t, s.Has("b"))
		assert.True(t, s.Has("c"))
	})

	t.Run("creates int set", func(t *testing.T) {
		s := set.New(1, 2, 3, 4, 5)
		assert.Equal(t, 5, len(s))
		assert.True(t, s.Has(3))
	})
}

func TestAdd(t *testing.T) {
	t.Run("adds item to empty set", func(t *testing.T) {
		s := set.New[string]()
		s.Add("item")
		assert.Equal(t, 1, len(s))
		assert.True(t, s.Has("item"))
	})

	t.Run("adds item to existing set", func(t *testing.T) {
		s := set.New("a", "b")
		s.Add("c")
		assert.Equal(t, 3, len(s))
		assert.True(t, s.Has("c"))
	})

	t.Run("adding duplicate item does not change size", func(t *testing.T) {
		s := set.New("a", "b")
		s.Add("a")
		assert.Equal(t, 2, len(s))
	})

	t.Run("adds multiple items sequentially", func(t *testing.T) {
		s := set.New[int]()
		s.Add(1)
		s.Add(2)
		s.Add(3)
		assert.Equal(t, 3, len(s))
	})
}

func TestRemove(t *testing.T) {
	t.Run("removes existing item", func(t *testing.T) {
		s := set.New("a", "b", "c")
		s.Remove("b")
		assert.Equal(t, 2, len(s))
		assert.True(t, s.Has("a"))
		assert.False(t, s.Has("b"))
		assert.True(t, s.Has("c"))
	})

	t.Run("removing non-existent item does nothing", func(t *testing.T) {
		s := set.New("a", "b")
		s.Remove("c")
		assert.Equal(t, 2, len(s))
		assert.True(t, s.Has("a"))
		assert.True(t, s.Has("b"))
	})

	t.Run("removes last item from set", func(t *testing.T) {
		s := set.New("only")
		s.Remove("only")
		assert.Equal(t, 0, len(s))
		assert.False(t, s.Has("only"))
	})

	t.Run("removes from empty set without panic", func(t *testing.T) {
		s := set.New[string]()
		assert.NotPanics(t, func() {
			s.Remove("item")
		})
	})
}

func TestHas(t *testing.T) {
	t.Run("returns true for existing item", func(t *testing.T) {
		s := set.New("a", "b", "c")
		assert.True(t, s.Has("b"))
	})

	t.Run("returns false for non-existent item", func(t *testing.T) {
		s := set.New("a", "b", "c")
		assert.False(t, s.Has("d"))
	})

	t.Run("returns false for empty set", func(t *testing.T) {
		s := set.New[string]()
		assert.False(t, s.Has("any"))
	})

	t.Run("works with int type", func(t *testing.T) {
		s := set.New(1, 2, 3)
		assert.True(t, s.Has(2))
		assert.False(t, s.Has(4))
	})
}

func TestSetOperations(t *testing.T) {
	t.Run("add then remove", func(t *testing.T) {
		s := set.New[string]()
		s.Add("item")
		assert.True(t, s.Has("item"))
		s.Remove("item")
		assert.False(t, s.Has("item"))
	})

	t.Run("re-add after remove", func(t *testing.T) {
		s := set.New("item")
		s.Remove("item")
		assert.False(t, s.Has("item"))
		s.Add("item")
		assert.True(t, s.Has("item"))
	})

	t.Run("set can be iterated", func(t *testing.T) {
		s := set.New("a", "b", "c")
		count := 0
		for range s {
			count++
		}
		assert.Equal(t, 3, count)
	})
}
