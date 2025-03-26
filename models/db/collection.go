package db

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DatabaseContextKey = "database"
)

var ErrNoDatabase = errors.New("no database in context")

func GetCollection(ctx context.Context, collectionName string) (*mongo.Collection, error) {
	db, ok := ctx.Value(DatabaseContextKey).(*mongo.Database)
	if !ok {
		return nil, errors.WithStack(ErrNoDatabase)
	}

	return db.Collection(collectionName), nil
}
