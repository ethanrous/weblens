// Package db provides MongoDB database operations with context integration and caching support.
package db

import (
	"context"
	"encoding/json"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// DatabaseContextKey is the context key for accessing the MongoDB database instance.
	DatabaseContextKey = "database"
	// CollectionContextKey is the context key for accessing the collection name.
	CollectionContextKey = "collection"
)

// ErrNoDatabase indicates that the context does not contain a database instance.
var ErrNoDatabase = wlerrors.New("context is not a DatabaseContext")

// ContextualizedCollection wraps a MongoDB collection with context for logging and caching.
type ContextualizedCollection[T any] struct {
	ctx        context.Context
	collection *mongo.Collection
}

// GetCollection returns the underlying MongoDB collection.
func (c *ContextualizedCollection[T]) GetCollection() *mongo.Collection {
	return c.collection
}

// InsertOne inserts a single document into the collection and invalidates the cache.
func (c *ContextualizedCollection[T]) InsertOne(_ context.Context, document any, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	log.FromContext(c.ctx).Trace().Func(func(e *zerolog.Event) {
		docStr, err := json.Marshal(document)
		if err == nil {
			if len(docStr) > 256 {
				docStr = append(docStr[:256], []byte("...")...)
			}

			e.CallerSkipFrame(4).Msgf("Insert on collection [%s] with document %v", c.collection.Name(), string(docStr))
		}
	})

	if config.GetConfig().DoCache {
		cache := context_mod.ToZ(c.ctx).GetCache(c.collection.Name())
		for _, key := range cache.ScanKeys() {
			cache.Delete(key)
		}
	}

	return c.collection.InsertOne(c.ctx, document, opts...)
}

// InsertMany inserts multiple documents into the collection.
func (c *ContextualizedCollection[T]) InsertMany(_ context.Context, documents []any, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	log.FromContext(c.ctx).Trace().Msgf("Insert many on collection [%s] with %d documents", c.collection.Name(), len(documents))

	return c.collection.InsertMany(c.ctx, documents, opts...)
}

// UpdateOne updates a single document matching the filter and invalidates the cache.
func (c *ContextualizedCollection[T]) UpdateOne(_ context.Context, filter, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	log.FromContext(c.ctx).Trace().Msgf("UpdateOne on collection [%s] with filter %v", c.collection.Name(), filter)

	if config.GetConfig().DoCache {
		cache := context_mod.ToZ(c.ctx).GetCache(c.collection.Name())
		for _, key := range cache.ScanKeys() {
			cache.Delete(key)
		}
	}

	res, err := c.collection.UpdateOne(c.ctx, filter, update, opts...)
	if err != nil {
		return res, err
	}

	if res.MatchedCount == 0 && res.UpsertedCount == 0 {
		return res, wlerrors.Errorf("no documents matched the filter: %v", filter)
	}

	return res, nil
}

// UpdateMany updates all documents matching the filter and invalidates the cache.
func (c *ContextualizedCollection[T]) UpdateMany(_ context.Context, filter, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	log.FromContext(c.ctx).Trace().Msgf("UpdateMany on collection [%s] with filter %v", c.collection.Name(), filter)

	if config.GetConfig().DoCache {
		cache := context_mod.ToZ(c.ctx).GetCache(c.collection.Name())
		for _, key := range cache.ScanKeys() {
			cache.Delete(key)
		}
	}

	res, err := c.collection.UpdateMany(c.ctx, filter, update, opts...)
	if err != nil {
		return res, err
	}

	return res, nil
}

// ReplaceOne replaces a single document matching the filter and invalidates the cache.
func (c *ContextualizedCollection[T]) ReplaceOne(_ context.Context, filter, replacement any, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	log.FromContext(c.ctx).Trace().Msgf("ReplaceOne on collection [%s] with filter %v", c.collection.Name(), filter)

	if config.GetConfig().DoCache {
		cache := context_mod.ToZ(c.ctx).GetCache(c.collection.Name())
		for _, key := range cache.ScanKeys() {
			cache.Delete(key)
		}
	}

	return c.collection.ReplaceOne(c.ctx, filter, replacement, opts...)
}

// FindOne finds a single document matching the filter with cache support.
func (c *ContextualizedCollection[T]) FindOne(_ context.Context, filter any, opts ...*options.FindOneOptions) Decoder[T] {
	if config.GetConfig().DoCache {
		cache := context_mod.ToZ(c.ctx).GetCache(c.collection.Name())

		filterStr, err := bson.Marshal(filter)
		if err != nil {
			return &errDecoder[T]{err}
		}

		v, ok := cache.Get(string(filterStr))
		if ok {
			log.FromContext(c.ctx).Trace().Msgf("Cache hit for collection [%s] with filter %v", c.collection.Name(), filter)

			return &decoder[T]{ctx: c.ctx, value: v}
		}

		log.FromContext(c.ctx).Trace().Msgf("FindOne on collection [%s] with filter %v", c.collection.Name(), filter)
	}

	ret := c.collection.FindOne(c.ctx, filter, opts...)

	return &mongoDecoder[T]{ctx: c.ctx, res: ret, filter: filter, col: c.collection.Name(), err: ret.Err()}
}

// FindOneAs finds a single document matching the filter and decodes it into the result type.
func (c *ContextualizedCollection[T]) FindOneAs(_ context.Context, filter any, opts ...*options.FindOneOptions) (T, error) {
	log.FromContext(c.ctx).Trace().Msgf("FindOneAs on collection [%s] with filter %v", c.collection.Name(), filter)

	var result T

	err := c.collection.FindOne(c.ctx, filter, opts...).Decode(&result)
	if err != nil {
		return result, WrapError(err, "find one as")
	}

	return result, nil
}

// Find finds all documents matching the filter and returns a cursor.
func (c *ContextualizedCollection[T]) Find(_ context.Context, filter any, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	log.FromContext(c.ctx).Trace().Msgf("Find on collection [%s] with filter %v", c.collection.Name(), filter)

	return c.collection.Find(c.ctx, filter, opts...)
}

// CountDocuments counts the number of documents matching the filter.
func (c *ContextualizedCollection[T]) CountDocuments(_ context.Context, filter any, opts ...*options.CountOptions) (int64, error) {
	log.FromContext(c.ctx).Trace().Msgf("CountDocuments on collection [%s] with filter %v", c.collection.Name(), filter)

	return c.collection.CountDocuments(c.ctx, filter, opts...)
}

// DeleteOne deletes a single document matching the filter.
func (c *ContextualizedCollection[T]) DeleteOne(_ context.Context, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return c.collection.DeleteOne(c.ctx, filter, opts...)
}

// DeleteMany deletes all documents matching the filter.
func (c *ContextualizedCollection[T]) DeleteMany(_ context.Context, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return c.collection.DeleteMany(c.ctx, filter, opts...)
}

// Aggregate executes an aggregation pipeline and returns a cursor with the results.
func (c *ContextualizedCollection[T]) Aggregate(_ context.Context, pipeline any, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	cursor, err := c.collection.Aggregate(c.ctx, pipeline, opts...)

	log.FromContext(c.ctx).Trace().Func(func(e *zerolog.Event) {
		if cursor == nil {
			e.Msgf("Aggregate on collection [%s] got nil cursor", c.collection.Name())

			return
		}

		e.Msgf("Aggregate on collection [%s] got %d results", c.collection.Name(), cursor.RemainingBatchLength())
	})

	return cursor, wlerrors.WithStack(err)
}

// Drop removes the entire collection and invalidates the cache.
func (c *ContextualizedCollection[T]) Drop(ctx context.Context) error {
	select {
	case <-c.ctx.Done():
	default:
		ctx = c.ctx
	}

	if config.GetConfig().DoCache {
		cache := context_mod.ToZ(c.ctx).GetCache(c.collection.Name())
		for _, key := range cache.ScanKeys() {
			cache.Delete(key)
		}
	}

	return c.collection.Drop(ctx)
}

func getDbFromContext(ctx context.Context) (*mongo.Database, error) {
	if ctx == nil {
		return nil, wlerrors.WithStack(ErrNoDatabase)
	}

	dbAny := ctx.Value(DatabaseContextKey)
	if dbAny == nil {
		return nil, wlerrors.WithStack(ErrNoDatabase)
	}

	db, ok := dbAny.(*mongo.Database)
	if ok {
		return db, nil
	}

	return nil, wlerrors.WithStack(ErrNoDatabase)
}

// GetCollection retrieves a contextualized collection from the database in the context.
func GetCollection[T any](ctx context.Context, collectionName string) (*ContextualizedCollection[T], error) {
	db, err := getDbFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ctxColName, ok := ctx.Value(CollectionContextKey).(string)
	if ok {
		collectionName = ctxColName
	}

	s := mongo.SessionFromContext(ctx)

	log.FromContext(ctx).Trace().Func(func(e *zerolog.Event) {
		if s != nil {
			e.CallerSkipFrame(4).Msgf("GetCollection [%s] with session %s", collectionName, s.ID())
		}
	})

	return &ContextualizedCollection[T]{ctx, db.Collection(collectionName)}, nil
}
