package db

import (
	"context"

	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/pkg/errors"
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
	return c.collection.InsertOne(c.ctx, document, opts...)
}

func (c *ContextualizedCollection) UpdateOne(_ context.Context, filter, update any, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return c.collection.UpdateOne(c.ctx, filter, update, opts...)
}

func (c *ContextualizedCollection) FindOne(_ context.Context, filter any, opts ...*options.FindOneOptions) *mongo.SingleResult {
	ret := c.collection.FindOne(c.ctx, filter, opts...)
	return ret
}

func (c *ContextualizedCollection) Find(_ context.Context, filter any, opts ...*options.FindOptions) (*mongo.Cursor, error) {
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

	dbCtx, ok := ctx.(context_mod.DatabaseContext)
	if ok {
		return dbCtx.Database(), nil
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

func hasTransaction(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	dbCtx, ok := ctx.(context_mod.DatabaseContext)
	if !ok {
		return false
	}

	seshCtx := dbCtx.GetMongoSession()
	return seshCtx != nil
}

func GetCollection(ctx context.Context, collectionName string) (*ContextualizedCollection, error) {
	db, err := getDbFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var seshCtx context.Context = ctx.(context_mod.DatabaseContext).GetMongoSession()
	if seshCtx == nil {
		seshCtx = ctx
	} else {
		ctx.(context_mod.ContextZ).Log().Debug().Msg("Using session context!")
	}

	// For tests
	ctxColName, ok := ctx.Value(CollectionContextKey).(string)
	if ok {
		return &ContextualizedCollection{seshCtx, db.Collection(ctxColName)}, nil
	}
	return &ContextualizedCollection{seshCtx, db.Collection(collectionName)}, nil
}

func WithTransaction(ctx context_mod.ContextZ, fn func(ctx context_mod.ContextZ) error) error {
	db, err := getDbFromContext(ctx)
	if err != nil {
		return err
	}

	if hasTransaction(ctx) {
		ctx.Log().Debug().Msg("Already in a transaction, skipping straight to callback function")
		return fn(ctx)
	}

	session, err := db.Client().StartSession()
	if err != nil {
		return errors.WithStack(err)
	}

	defer session.EndSession(ctx)

	ctx.Log().Info().Msg("Starting transaction!")
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		// TODO: copy the context to avoid duplicate transactions
		ctx.(context_mod.DatabaseContext).WithMongoSession(sessCtx)
		err := fn(ctx)
		if err != nil {
			return nil, err
		}

		ctx.Log().Debug().Msg("Transaction complete, committing!")
		err = sessCtx.CommitTransaction(sessCtx)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return nil, nil
	})

	return err
}
