package db

import (
	"context"
	"reflect"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrDecodeNotPointer indicates that a decode operation was attempted with a non-pointer value.
var ErrDecodeNotPointer = wlerrors.New("decode: value must be a pointer")

// Decoder provides an interface for decoding database results into typed values.
type Decoder[T any] interface {
	Decode(v T) error
	Err() error
}

type decoder[T any] struct {
	ctx   context.Context
	value any
}

func (d *decoder[T]) Decode(v T) error {
	if d.value == nil {
		return mongo.ErrNoDocuments
	}

	rval := reflect.ValueOf(v)
	if rval.Kind() != reflect.Pointer {
		return ErrDecodeNotPointer
	}

	rval.Elem().Set(reflect.ValueOf(d.value).Elem())

	return nil
}

func (d *decoder[T]) Err() error {
	return nil
}

type mongoDecoder[T any] struct {
	ctx    context.Context
	res    *mongo.SingleResult
	filter any
	col    string
	err    error
}

func (d *mongoDecoder[T]) Decode(v T) error {
	if d.err != nil {
		if wlerrors.Is(d.err, mongo.ErrNoDocuments) {
			err := d.cacheResult(nil)
			if err != nil {
				return wlerrors.Errorf("failed to negative cache result: %v", err)
			}
		}

		return d.err
	}

	err := d.res.Decode(v)
	if err != nil {
		log.GlobalLogger().Debug().Msgf("decode error %v", err)

		return err
	}

	err = d.cacheResult(v)
	if err != nil {
		return wlerrors.Errorf("failed to cache result: %v", err)
	}

	return nil
}

func (d *mongoDecoder[T]) Err() error {
	return d.err
}

type errDecoder[T any] struct {
	err error
}

func (d *errDecoder[T]) Decode(T) error {
	return d.err
}

func (d *errDecoder[T]) Err() error {
	return d.err
}

func (d *mongoDecoder[T]) cacheResult(v any) error {
	if config.GetConfig().DoCache {
		cache := context_mod.ToZ(d.ctx).GetCache(d.col)

		m, ok := d.filter.(bson.M)
		if !ok {
			return wlerrors.New("filter is not bson.M")
		}

		if len(m) != 1 {
			return nil
		}

		filterString, err := bson.Marshal(m)
		if err != nil {
			return wlerrors.Wrap(err, "failed to marshal filter")
		}

		cache.Set(string(filterString), v)
	}

	return nil
}
