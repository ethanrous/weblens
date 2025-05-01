package db

import (
	"context"

	"github.com/ethanrous/weblens/modules/config"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DatabaseContextKey   = "database"
	CollectionContextKey = "collection"
)

var ErrNoDatabase = errors.New("context is not a DatabaseContext")

type ContextualizedCollection struct {
	ctx        context.Context
	collection *mongo.Collection
}

func (c *ContextualizedCollection) InsertOne(_ context.Context, document any, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	ctx := context_mod.ToZ(c.ctx)
	ctx.Log().Trace().Msgf("Insert on collection [%s] with document %v", c.collection.Name(), document)

	if config.GetConfig().DoCache {
		cache := ctx.GetCache(c.collection.Name())
		for _, key := range cache.ScanKeys() {
			cache.Delete(key)
		}
	}

	return c.collection.InsertOne(c.ctx, document, opts...)
}

func (c *ContextualizedCollection) InsertMany(_ context.Context, documents []any, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	context_mod.ToZ(c.ctx).Log().Trace().Msgf("Insert many on collection [%s] with %d documents", c.collection.Name(), len(documents))

	return c.collection.InsertMany(c.ctx, documents, opts...)
}

func (c *ContextualizedCollection) UpdateOne(_ context.Context, filter, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	ctx := context_mod.ToZ(c.ctx)
	ctx.Log().Trace().Msgf("UpdateOne on collection [%s] with filter %v", c.collection.Name(), filter)

	if config.GetConfig().DoCache {
		cache := ctx.GetCache(c.collection.Name())
		for _, key := range cache.ScanKeys() {
			cache.Delete(key)
		}
	}

	return c.collection.UpdateOne(c.ctx, filter, update, opts...)
}

func (c *ContextualizedCollection) FindOne(_ context.Context, filter any, opts ...*options.FindOneOptions) Decoder {
	if config.GetConfig().DoCache {
		cache := context_mod.ToZ(c.ctx).GetCache(c.collection.Name())

		filterStr, err := bson.Marshal(filter)
		if err != nil {
			return &errDecoder{err}
		}

		v, ok := cache.Get(string(filterStr))
		if ok {
			context_mod.ToZ(c.ctx).Log().Trace().Msgf("Cache hit for collection [%s] with filter %v", c.collection.Name(), filter)

			return &decoder{ctx: c.ctx, value: v}
		}

		context_mod.ToZ(c.ctx).Log().Trace().Msgf("FindOne on collection [%s] with filter %v", c.collection.Name(), filter)
	}

	ret := c.collection.FindOne(c.ctx, filter, opts...)

	return &mongoDecoder{ctx: c.ctx, res: ret, filter: filter, col: c.collection.Name()}
}

func (c *ContextualizedCollection) Find(_ context.Context, filter any, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	context_mod.ToZ(c.ctx).Log().Trace().Msgf("Find on collection [%s] with filter %v", c.collection.Name(), filter)

	return c.collection.Find(c.ctx, filter, opts...)
}

func (c *ContextualizedCollection) DeleteOne(_ context.Context, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return c.collection.DeleteOne(c.ctx, filter, opts...)
}

func (c *ContextualizedCollection) DeleteMany(_ context.Context, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return c.collection.DeleteMany(c.ctx, filter, opts...)
}

func (c *ContextualizedCollection) Aggregate(_ context.Context, pipeline any, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	return c.collection.Aggregate(c.ctx, pipeline, opts...)
}

func (c *ContextualizedCollection) Drop(_ context.Context) error {
	return c.collection.Drop(c.ctx)
}

func getDbFromContext(ctx context.Context) (*mongo.Database, error) {
	if ctx == nil {
		return nil, errors.WithStack(ErrNoDatabase)
	}

	dbAny := ctx.Value(DatabaseContextKey)
	if dbAny == nil {
		return nil, errors.WithStack(ErrNoDatabase)
	}

	db, ok := dbAny.(*mongo.Database)
	if ok {
		return db, nil
	}

	return nil, errors.WithStack(ErrNoDatabase)
}

func GetCollection(ctx context.Context, collectionName string) (*ContextualizedCollection, error) {
	db, err := getDbFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ctxColName, ok := ctx.Value(CollectionContextKey).(string)
	if ok {
		collectionName = ctxColName
	}

	s := mongo.SessionFromContext(ctx)

	context_mod.ToZ(ctx).Log().Trace().Func(func(e *zerolog.Event) {
		if s == nil {
			e.Msgf("GetCollection [%s] without session", collectionName)
		} else {
			e.Msgf("GetCollection [%s] with session %s", collectionName, s.ID())
		}
	})

	return &ContextualizedCollection{ctx, db.Collection(collectionName)}, nil
}
