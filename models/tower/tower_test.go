package tower

import (
	"context"
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BSON keys
const (
	KeyTowerId     = "towerId"
	KeyTowerRole   = "towerRole"
	KeyIsThisTower = "isThisTower"
	KeyCreatedBy   = "createdBy"
)

func TestCreateLocal(t *testing.T) {
	ctx := context.Background()
	mongodb, err := db.ConnectToMongo(ctx, config.GetMongoDBUri(), "weblensTestDB")
	if err != nil {
		t.Error(err)
	}
	ctx = context.WithValue(context.Background(), db.DatabaseContextKey, mongodb)

	t.Run("success", func(t *testing.T) {
		ctx := context.WithValue(ctx, db.CollectionContextKey, t.Name())
		tower, err := CreateLocal(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, tower)
		assert.Equal(t, InitTowerRole, tower.Role)
		assert.True(t, tower.IsThisTower)
	})
}

func TestCreateTower(t *testing.T) {
	ctx := context.Background()
	mongodb, err := db.ConnectToMongo(ctx, config.GetMongoDBUri(), "weblensTestDB")
	if err != nil {
		t.Error(err)
	}
	ctx = context.WithValue(context.Background(), db.DatabaseContextKey, mongodb)

	t.Run("success", func(t *testing.T) {
		ctx := context.WithValue(ctx, db.CollectionContextKey, t.Name())
		tower := &Instance{
			DbId:    primitive.NewObjectID(),
			TowerId: primitive.NewObjectID().Hex(),
			Role:    CoreTowerRole,
		}

		err := CreateTower(ctx, tower)

		assert.NoError(t, err)
	})
}

func TestGetTowerById(t *testing.T) {
	ctx := context.Background()
	mongodb, err := db.ConnectToMongo(ctx, config.GetMongoDBUri(), "weblensTestDB")
	if err != nil {
		t.Error(err)
	}
	ctx = context.WithValue(context.Background(), db.DatabaseContextKey, mongodb)

	t.Run("found", func(t *testing.T) {
		ctx := context.WithValue(ctx, db.CollectionContextKey, t.Name())
		towerId := primitive.NewObjectID().Hex()
		// Insert a mock tower document
		_, err := mongodb.Collection(t.Name()).InsertOne(ctx, bson.D{
			{Key: KeyTowerId, Value: towerId},
			{Key: KeyTowerRole, Value: string(CoreTowerRole)},
		})
		assert.NoError(t, err)

		tower, err := GetTowerById(ctx, towerId)

		assert.NoError(t, err)
		assert.NotNil(t, tower)
		assert.Equal(t, towerId, tower.TowerId)
	})

	t.Run("not found", func(t *testing.T) {
		ctx := context.WithValue(ctx, db.CollectionContextKey, t.Name())
		tower, err := GetTowerById(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, tower)
		assert.Equal(t, ErrTowerNotFound, err)
	})
}

func TestGetLocal(t *testing.T) {
	ctx := context.Background()
	mongodb, err := db.ConnectToMongo(ctx, config.GetMongoDBUri(), "weblensTestDB")
	if err != nil {
		t.Error(err)
	}
	ctx = context.WithValue(context.Background(), db.DatabaseContextKey, mongodb)

	t.Run("found", func(t *testing.T) {
		ctx := context.WithValue(ctx, db.CollectionContextKey, t.Name())
		// Insert a mock tower document
		_, err := mongodb.Collection(t.Name()).InsertOne(ctx, bson.D{
			{Key: KeyIsThisTower, Value: true},
			{Key: KeyTowerRole, Value: string(CoreTowerRole)},
		})
		assert.NoError(t, err)

		tower, err := GetLocal(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, tower)
		assert.True(t, tower.IsThisTower)
	})

	t.Run("not found", func(t *testing.T) {
		ctx := context.WithValue(ctx, db.CollectionContextKey, t.Name())
		tower, err := GetLocal(ctx)

		assert.Error(t, err)
		assert.Nil(t, tower)
	})
}

func TestSetLastBackup(t *testing.T) {
	ctx := context.Background()
	mongodb, err := db.ConnectToMongo(ctx, config.GetMongoDBUri(), "weblensTestDB")
	if err != nil {
		t.Error(err)
	}
	ctx = context.WithValue(context.Background(), db.DatabaseContextKey, mongodb)

	t.Run("success", func(t *testing.T) {
		ctx := context.WithValue(ctx, db.CollectionContextKey, t.Name())
		// Insert a mock tower document
		towerId := "someTowerId"
		_, err := mongodb.Collection(t.Name()).InsertOne(ctx, bson.D{
			{Key: KeyTowerId, Value: towerId},
		})
		assert.NoError(t, err)

		err = SetLastBackup(ctx, towerId, time.Now())

		assert.NoError(t, err)
	})
}

func TestGetAllTowersByTowerId(t *testing.T) {
	ctx := context.Background()
	mongodb, err := db.ConnectToMongo(ctx, config.GetMongoDBUri(), "weblensTestDB")
	if err != nil {
		t.Error(err)
	}
	ctx = context.WithValue(context.Background(), db.DatabaseContextKey, mongodb)

	t.Run("found", func(t *testing.T) {
		ctx := context.WithValue(ctx, db.CollectionContextKey, t.Name())
		towerId := "creatorTowerId"
		// Insert a mock tower document
		_, err := mongodb.Collection(t.Name()).InsertOne(ctx, bson.D{
			{Key: KeyCreatedBy, Value: towerId},
			{Key: KeyTowerRole, Value: string(CoreTowerRole)},
		})
		assert.NoError(t, err)

		towers, err := GetAllTowersByTowerId(ctx, towerId)

		assert.NoError(t, err)
		assert.NotNil(t, towers)
		assert.Len(t, towers, 1)
	})

	t.Run("not found", func(t *testing.T) {
		ctx := context.WithValue(ctx, db.CollectionContextKey, t.Name())
		towers, err := GetAllTowersByTowerId(ctx, "nonexistent")

		assert.NoError(t, err)
		assert.Empty(t, towers)
	})
}
