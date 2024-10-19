package mock

import (
	"context"

	"github.com/ethanrous/weblens/database"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ database.MongoCollection = (*MockFailMongoCol)(nil)

type MockFailMongoCol struct {
	RealCol    *mongo.Collection
	InsertFail bool
	FindFail   bool
	UpdateFail bool
	DeleteFail bool

	Inserts []any
	Finds   []any
	Updates []any
	Deletes []any
}

func (fc *MockFailMongoCol) InsertOne(
	ctx context.Context, document interface{}, opts ...*options.InsertOneOptions,
) (*mongo.InsertOneResult, error) {
	if fc.InsertFail {
		return nil, mongo.ErrNoDocuments
	}

	return fc.RealCol.InsertOne(ctx, document, opts...)
}

func (fc *MockFailMongoCol) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (
	cur *mongo.Cursor, err error,
) {
	if fc.FindFail {
		return nil, mongo.ErrNoDocuments
	}

	return fc.RealCol.Find(ctx, filter, opts...)
}

func (fc *MockFailMongoCol) UpdateOne(
	ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions,
) (*mongo.UpdateResult, error) {
	if fc.UpdateFail {
		return nil, mongo.ErrNoDocuments
	}
	return fc.RealCol.UpdateOne(ctx, filter, update, opts...)
}

func (fc *MockFailMongoCol) DeleteOne(
	ctx context.Context, filter interface{}, opts ...*options.DeleteOptions,
) (*mongo.DeleteResult, error) {
	if fc.DeleteFail {
		return nil, mongo.ErrNoDocuments
	}
	return fc.RealCol.DeleteOne(ctx, filter, opts...)
}

func (fc *MockFailMongoCol) DeleteMany(
	ctx context.Context, filter interface{}, opts ...*options.DeleteOptions,
) (*mongo.DeleteResult, error) {
	if fc.DeleteFail {
		return nil, mongo.ErrNoDocuments
	}
	return fc.RealCol.DeleteMany(ctx, filter, opts...)
}
