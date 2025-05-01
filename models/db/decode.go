package db

import (
	"context"
	"reflect"

	"github.com/ethanrous/weblens/modules/config"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrDecodeNotPointer = errors.New("decode: value must be a pointer")

type Decoder interface {
	Decode(v any) error
}

type decoder struct {
	ctx   context.Context
	value any
}

func (d *decoder) Decode(v any) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() != reflect.Pointer {
		return ErrDecodeNotPointer
	}

	rval.Elem().Set(reflect.ValueOf(d.value).Elem())

	return nil
}

type mongoDecoder struct {
	ctx    context.Context
	res    *mongo.SingleResult
	filter any
	col    string
	err    error
}

func (d *mongoDecoder) Decode(v any) error {
	if d.err != nil {
		return d.err
	}

	err := d.res.Decode(v)
	if err != nil {
		return err
	}

	if config.GetConfig().DoCache {
		cache := context_mod.ToZ(d.ctx).GetCache(d.col)

		m, ok := d.filter.(bson.M)
		if !ok {
			return errors.New("filter is not bson.M")
		}

		if len(m) != 1 {
			return nil
		}

		filterString, err := bson.Marshal(m)
		if err != nil {
			return errors.Wrap(err, "failed to marshal filter")
		}

		cache.Set(string(filterString), v)
	}

	return nil
}

type errDecoder struct {
	err error
}

func (d *errDecoder) Decode(any) error {
	return d.err
}
