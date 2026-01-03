
package option_test

import (
	"encoding/json"
	"testing"

	"github.com/ethanrous/weblens/modules/option"
	"github.com/stretchr/testify/assert"
)

func TestOf(t *testing.T) {
	t.Run("creates option with string value", func(t *testing.T) {
		opt := option.Of("hello")
		assert.True(t, opt.Has())

		val, ok := opt.Get()
		assert.True(t, ok)
		assert.Equal(t, "hello", val)
	})

	t.Run("creates option with int value", func(t *testing.T) {
		opt := option.Of(42)
		assert.True(t, opt.Has())

		val, ok := opt.Get()
		assert.True(t, ok)
		assert.Equal(t, 42, val)
	})

	t.Run("creates option with zero value", func(t *testing.T) {
		opt := option.Of(0)
		assert.True(t, opt.Has())

		val, ok := opt.Get()
		assert.True(t, ok)
		assert.Equal(t, 0, val)
	})

	t.Run("creates option with nil pointer", func(t *testing.T) {
		var ptr *string
		opt := option.Of(ptr)
		assert.True(t, opt.Has())

		val, ok := opt.Get()
		assert.True(t, ok)
		assert.Nil(t, val)
	})
}

func TestZero(t *testing.T) {
	t.Run("creates empty string option", func(t *testing.T) {
		opt := option.Zero[string]()
		assert.False(t, opt.Has())

		val, ok := opt.Get()
		assert.False(t, ok)
		assert.Equal(t, "", val)
	})

	t.Run("creates empty int option", func(t *testing.T) {
		opt := option.Zero[int]()
		assert.False(t, opt.Has())

		val, ok := opt.Get()
		assert.False(t, ok)
		assert.Equal(t, 0, val)
	})
}

func TestHas(t *testing.T) {
	t.Run("returns true for present value", func(t *testing.T) {
		opt := option.Of("test")
		assert.True(t, opt.Has())
	})

	t.Run("returns false for absent value", func(t *testing.T) {
		opt := option.Zero[string]()
		assert.False(t, opt.Has())
	})
}

func TestGet(t *testing.T) {
	t.Run("returns value and true for present", func(t *testing.T) {
		opt := option.Of("value")
		val, ok := opt.Get()
		assert.True(t, ok)
		assert.Equal(t, "value", val)
	})

	t.Run("returns zero and false for absent", func(t *testing.T) {
		opt := option.Zero[int]()
		val, ok := opt.Get()
		assert.False(t, ok)
		assert.Equal(t, 0, val)
	})
}

func TestGetOr(t *testing.T) {
	t.Run("returns value when present", func(t *testing.T) {
		opt := option.Of("actual")
		val := opt.GetOr("fallback")
		assert.Equal(t, "actual", val)
	})

	t.Run("returns fallback when absent", func(t *testing.T) {
		opt := option.Zero[string]()
		val := opt.GetOr("fallback")
		assert.Equal(t, "fallback", val)
	})

	t.Run("returns zero value fallback", func(t *testing.T) {
		opt := option.Zero[int]()
		val := opt.GetOr(0)
		assert.Equal(t, 0, val)
	})
}

func TestSet(t *testing.T) {
	t.Run("sets value on empty option", func(t *testing.T) {
		opt := option.Zero[string]()
		assert.False(t, opt.Has())

		newOpt := opt.Set("new value")
		assert.True(t, newOpt.Has())

		val, ok := newOpt.Get()
		assert.True(t, ok)
		assert.Equal(t, "new value", val)
	})

	t.Run("overwrites existing value", func(t *testing.T) {
		opt := option.Of("old")
		newOpt := opt.Set("new")

		val, ok := newOpt.Get()
		assert.True(t, ok)
		assert.Equal(t, "new", val)
	})
}

func TestMarshalJSON(t *testing.T) {
	t.Run("marshals present string value", func(t *testing.T) {
		opt := option.Of("hello")
		data, err := json.Marshal(opt)
		assert.NoError(t, err)
		assert.Equal(t, `"hello"`, string(data))
	})

	t.Run("marshals present int value", func(t *testing.T) {
		opt := option.Of(42)
		data, err := json.Marshal(opt)
		assert.NoError(t, err)
		assert.Equal(t, "42", string(data))
	})

	t.Run("marshals absent value as null", func(t *testing.T) {
		opt := option.Zero[string]()
		data, err := json.Marshal(opt)
		assert.NoError(t, err)
		assert.Equal(t, "null", string(data))
	})

	t.Run("marshals struct value", func(t *testing.T) {
		type testStruct struct {
			Name string `json:"name"`
		}

		opt := option.Of(testStruct{Name: "test"})
		data, err := json.Marshal(opt)
		assert.NoError(t, err)
		assert.Equal(t, `{"name":"test"}`, string(data))
	})

	t.Run("marshals slice value", func(t *testing.T) {
		opt := option.Of([]int{1, 2, 3})
		data, err := json.Marshal(opt)
		assert.NoError(t, err)
		assert.Equal(t, "[1,2,3]", string(data))
	})
}
