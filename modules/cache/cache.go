package cache

import (
	"context"

	"github.com/ethanrous/weblens/modules/errors"
	"github.com/viccon/sturdyc"
)

var ErrNotCacher = errors.Errorf("context does not implement cacher interface")
var ErrNoCache = errors.Errorf("cache not found")

type cacher interface {
	GetCache(col string) *sturdyc.Client[any]
}

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

func SetOne[T any](ctx context.Context, cacheKey, itemKey string, item T) error {
	cacheCtx, ok := ctx.(cacher)
	if !ok {
		return errors.WithStack(ErrNotCacher)
	}

	cache := cacheCtx.GetCache(cacheKey)
	if cache == nil {
		return errors.WithStack(ErrNoCache)
	}

	cache.Set(itemKey, item)

	return nil
}

func Drop(ctx context.Context, cacheKey string) error {
	cacheCtx, ok := ctx.(cacher)
	if !ok {
		return errors.WithStack(ErrNotCacher)
	}

	cache := cacheCtx.GetCache(cacheKey)
	if cache == nil {
		return errors.WithStack(ErrNoCache)
	}

	for _, k := range cache.ScanKeys() {
		cache.Delete(k)
	}

	return nil
}
