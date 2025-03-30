package context

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type DatabaseContext interface {
	context.Context
	Database() *mongo.Database
}
