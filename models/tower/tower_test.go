package tower

import (
	"context"
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/tests"
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

func initLocalTower(t *testing.T, ctx context.Context) Instance {
	i, err := CreateLocal(ctx)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	return i
}

func TestCreateLocal(t *testing.T) {
	t.Parallel()

	mongodb, err := db.ConnectToMongo(t.Context(), config.GetMongoDBUri(), "weblensTestDB")
	if err != nil {
		t.Error(err)
	}

	t.Run("success", func(t *testing.T) {
		defer tests.Recover(t)
		ctx := context.WithValue(t.Context(), db.DatabaseContextKey, mongodb)
		tower, err := CreateLocal(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, tower)
		assert.Equal(t, RoleInit, tower.Role)
		assert.True(t, tower.IsThisTower)
	})
}

func TestCreateTower(t *testing.T) {
	defer tests.Recover(t)
	t.Parallel()
	log.NewZeroLogger()

	mongodb, err := db.ConnectToMongo(t.Context(), config.GetMongoDBUri(), "weblensTestDB")
	if err != nil {
		t.Error(err)
	}

	t.Run("success", func(t *testing.T) {
		defer tests.Recover(t)
		ctx := context.WithValue(t.Context(), db.DatabaseContextKey, mongodb)

		initLocalTower(t, ctx)

		tower := &Instance{
			Name:    "Test Tower",
			DbId:    primitive.NewObjectID(),
			TowerId: primitive.NewObjectID().Hex(),
			Role:    RoleCore,
		}

		err := SaveTower(ctx, tower)

		assert.NoError(t, err)
	})

	t.Run("bad tower", func(t *testing.T) {
		defer tests.Recover(t)
		ctx := context.WithValue(t.Context(), db.DatabaseContextKey, mongodb)

		initLocalTower(t, ctx)

		tower := &Instance{
			DbId:    primitive.NewObjectID(),
			TowerId: primitive.NewObjectID().Hex(),
			Role:    RoleCore,
		}

		err := SaveTower(ctx, tower)

		assert.Error(t, err)
	})
}

func TestGetTowerById(t *testing.T) {
	mongodb, err := db.ConnectToMongo(t.Context(), config.GetMongoDBUri(), "weblensTestDB")
	if err != nil {
		t.Error(err)
	}

	t.Run("found", func(t *testing.T) {
		defer tests.Recover(t)
		ctx := context.WithValue(t.Context(), db.DatabaseContextKey, mongodb)

		local := initLocalTower(t, ctx)

		tower, err := GetTowerById(ctx, local.TowerId)

		assert.NoError(t, err)
		assert.NotNil(t, tower)
		assert.Equal(t, local.TowerId, tower.TowerId)
	})

	t.Run("not found", func(t *testing.T) {
		defer tests.Recover(t)
		ctx := context.WithValue(t.Context(), db.DatabaseContextKey, mongodb)
		tower, err := GetTowerById(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Equal(t, Instance{}, tower)
		assert.Equal(t, ErrTowerNotFound, err)
	})
}

func TestGetLocal(t *testing.T) {
	mongodb, err := db.ConnectToMongo(t.Context(), config.GetMongoDBUri(), "weblensTestDB")
	if err != nil {
		t.Error(err)
	}

	t.Run("found", func(t *testing.T) {
		defer tests.Recover(t)
		ctx := context.WithValue(t.Context(), db.DatabaseContextKey, mongodb)

		initLocalTower(t, ctx)

		tower, err := GetLocal(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, tower)
		assert.True(t, tower.IsThisTower)
	})

	t.Run("not found", func(t *testing.T) {
		defer tests.Recover(t)
		ctx := context.WithValue(t.Context(), db.DatabaseContextKey, mongodb)
		mongodb.Collection(TowerCollectionKey).Drop(ctx)

		tower, err := GetLocal(ctx)

		assert.Error(t, err)
		assert.Equal(t, Instance{}, tower)
	})
}

func TestSetLastBackup(t *testing.T) {
	mongodb, err := db.ConnectToMongo(t.Context(), config.GetMongoDBUri(), "weblensTestDB")
	if err != nil {
		t.Error(err)
	}

	t.Run("success", func(t *testing.T) {
		defer tests.Recover(t)
		ctx := context.WithValue(t.Context(), db.DatabaseContextKey, mongodb)
		initLocalTower(t, ctx)
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

	mongodb, err := db.ConnectToMongo(t.Context(), config.GetMongoDBUri(), "weblensTestDB")
	if err != nil {
		t.Error(err)
	}

	t.Run("found", func(t *testing.T) {
		defer tests.Recover(t)
		ctx := context.WithValue(t.Context(), db.DatabaseContextKey, mongodb)

		local := initLocalTower(t, ctx)

		towerToFind := []Instance{
			{
				Name:      "Test Tower",
				CreatedBy: local.TowerId,
				TowerId:   "towerId1",
				Role:      RoleCore,
			},
			{
				Name:      "Test Tower2",
				CreatedBy: local.TowerId,
				TowerId:   "towerId2",
				Role:      RoleCore,
			},
		}

		// Insert mock tower documents
		for _, tower := range towerToFind {
			err = SaveTower(ctx, &tower)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
		}

		towers, err := GetAllTowersByTowerId(ctx, local.TowerId)

		assert.NoError(t, err)
		assert.NotNil(t, towers)
		assert.Len(t, towers, 2)
	})

	t.Run("not found", func(t *testing.T) {
		defer tests.Recover(t)
		ctx := context.WithValue(t.Context(), db.DatabaseContextKey, mongodb)
		towers, err := GetAllTowersByTowerId(ctx, "nonexistent")

		assert.NoError(t, err)
		assert.Empty(t, towers)
	})
}
