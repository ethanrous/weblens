// Package cache provides caching utilities using Sturdyc for the Weblens system.
package cache

import (
	"context"

	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/viccon/sturdyc"
)

// ErrNotCacher is returned when the context does not implement the cacher interface.
var ErrNotCacher = wlerrors.Errorf("context does not implement cacher interface")

// ErrNoCache is returned when the requested cache is not found.
var ErrNoCache = wlerrors.Errorf("cache not found")

type cacher interface {
	GetCache(col string) *sturdyc.Client[any]
}

// GetOneAs retrieves an item from the cache with the given cache key and item key, returning it as type T.
func GetOneAs[T any](ctx context.Context, cacheKey, itemKey string) (T, bool) {
	var zero T

	cacheCtx, ok := ctx.(cacher)
	if !ok {
		return zero, false
	}

	item, exists := cacheCtx.GetCache(cacheKey).Get(itemKey)
	if !exists {
		return zero, false
	}

	tItem, ok := item.(T)
	if !ok {
		return zero, false
	}

	return tItem, true
}

// SetOne stores an item in the cache with the given cache key and item key.
func SetOne[T any](ctx context.Context, cacheKey, itemKey string, item T) error {
	cacheCtx, ok := ctx.(cacher)
	if !ok {
		return wlerrors.WithStack(ErrNotCacher)
	}

	cache := cacheCtx.GetCache(cacheKey)
	if cache == nil {
		return wlerrors.WithStack(ErrNoCache)
	}

	cache.Set(itemKey, item)

	return nil
}

// Drop removes all items from the cache with the given cache key.
func Drop(ctx context.Context, cacheKey string) error {
	cacheCtx, ok := ctx.(cacher)
	if !ok {
		return wlerrors.WithStack(ErrNotCacher)
	}

	cache := cacheCtx.GetCache(cacheKey)
	if cache == nil {
		return wlerrors.WithStack(ErrNoCache)
	}

	for _, k := range cache.ScanKeys() {
		cache.Delete(k)
	}

	return nil
}
