package db

import (
	"context"
	"encoding/json"

	"github.com/ethanrous/weblens/modules/config"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
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

func (c *ContextualizedCollection) GetCollection() *mongo.Collection {
	return c.collection
}

func (c *ContextualizedCollection) InsertOne(_ context.Context, document any, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
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

func (c *ContextualizedCollection) InsertMany(_ context.Context, documents []any, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	log.FromContext(c.ctx).Trace().Msgf("Insert many on collection [%s] with %d documents", c.collection.Name(), len(documents))

	return c.collection.InsertMany(c.ctx, documents, opts...)
}

func (c *ContextualizedCollection) UpdateOne(_ context.Context, filter, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
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
		return res, errors.Errorf("no documents matched the filter: %v", filter)
	}

	return res, nil
}

func (c *ContextualizedCollection) ReplaceOne(_ context.Context, filter, replacement any, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	log.FromContext(c.ctx).Trace().Msgf("ReplaceOne on collection [%s] with filter %v", c.collection.Name(), filter)

	if config.GetConfig().DoCache {
		cache := context_mod.ToZ(c.ctx).GetCache(c.collection.Name())
		for _, key := range cache.ScanKeys() {
			cache.Delete(key)
		}
	}

	return c.collection.ReplaceOne(c.ctx, filter, replacement, opts...)
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
			log.FromContext(c.ctx).Trace().Msgf("Cache hit for collection [%s] with filter %v", c.collection.Name(), filter)

			return &decoder{ctx: c.ctx, value: v}
		}

		log.FromContext(c.ctx).Trace().Msgf("FindOne on collection [%s] with filter %v", c.collection.Name(), filter)
	}

	ret := c.collection.FindOne(c.ctx, filter, opts...)

	return &mongoDecoder{ctx: c.ctx, res: ret, filter: filter, col: c.collection.Name(), err: ret.Err()}
}

func (c *ContextualizedCollection) Find(_ context.Context, filter any, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	log.FromContext(c.ctx).Trace().Msgf("Find on collection [%s] with filter %v", c.collection.Name(), filter)

	return c.collection.Find(c.ctx, filter, opts...)
}

func (c *ContextualizedCollection) CountDocuments(_ context.Context, filter any, opts ...*options.CountOptions) (int64, error) {
	log.FromContext(c.ctx).Trace().Msgf("CountDocuments on collection [%s] with filter %v", c.collection.Name(), filter)

	return c.collection.CountDocuments(c.ctx, filter, opts...)
}

func (c *ContextualizedCollection) DeleteOne(_ context.Context, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return c.collection.DeleteOne(c.ctx, filter, opts...)
}

func (c *ContextualizedCollection) DeleteMany(_ context.Context, filter any, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return c.collection.DeleteMany(c.ctx, filter, opts...)
}

func (c *ContextualizedCollection) Aggregate(_ context.Context, pipeline any, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	cursor, err := c.collection.Aggregate(c.ctx, pipeline, opts...)

	log.FromContext(c.ctx).Trace().Msgf("Aggregate on collection [%s] got %d results", c.collection.Name(), cursor.RemainingBatchLength())

	return cursor, errors.WithStack(err)
}

func (c *ContextualizedCollection) Drop(ctx context.Context) error {
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

	log.FromContext(ctx).Trace().Func(func(e *zerolog.Event) {
		if s == nil {
			e.CallerSkipFrame(4).Msgf("GetCollection [%s] without session", collectionName)
		} else {
			e.CallerSkipFrame(4).Msgf("GetCollection [%s] with session %s", collectionName, s.ID())
		}
	})

	return &ContextualizedCollection{ctx, db.Collection(collectionName)}, nil
}
