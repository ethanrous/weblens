package context

import (
	"context"

	"github.com/viccon/sturdyc"
	"go.mongodb.org/mongo-driver/mongo"
)

type DatabaseContext interface {
	context.Context
	Database() *mongo.Database
	GetCache(cacheName string) *sturdyc.Client[any]
}
