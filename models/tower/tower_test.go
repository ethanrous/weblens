package tower_test

import (
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	. "github.com/ethanrous/weblens/models/tower"
)

const (
	testTowerName = "Test Tower"
)

func TestTower_Creation(t *testing.T) {
	ctx := db.SetupTestDB(t, TowerCollectionKey)

	t.Run("CreateLocal", func(t *testing.T) {
		tower, err := CreateLocal(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, tower)
		assert.Equal(t, RoleInit, tower.Role)
		assert.True(t, tower.IsThisTower)

		// Verify tower was saved to database
		savedTower, err := GetTowerById(ctx, tower.TowerId)
		assert.NoError(t, err)
		assert.Equal(t, tower.TowerId, savedTower.TowerId)
	})

	t.Run("CreateTower", func(t *testing.T) {
		tower := &Instance{
			TowerId: primitive.NewObjectID().Hex(),
			Name:    testTowerName,
			Role:    RoleCore,
		}

		err := SaveTower(ctx, tower)
		assert.NoError(t, err)

		// Verify tower was saved
		savedTower, err := GetTowerById(ctx, tower.TowerId)
		assert.NoError(t, err)
		assert.Equal(t, tower.TowerId, savedTower.TowerId)
		assert.Equal(t, testTowerName, savedTower.Name)
	})

	t.Run("CreateInvalidTower", func(t *testing.T) {
		tower := &Instance{
			TowerId: primitive.NewObjectID().Hex(),
			Role:    RoleCore,
			// Missing Name
		}

		err := SaveTower(ctx, tower)
		assert.Error(t, err)
	})
}

func TestTower_Retrieval(t *testing.T) {
	ctx := db.SetupTestDB(t, TowerCollectionKey)

	t.Run("GetTowerById", func(t *testing.T) {
		tower := &Instance{
			TowerId: primitive.NewObjectID().Hex(),
			Name:    testTowerName,
			Role:    RoleCore,
		}

		err := SaveTower(ctx, tower)
		require.NoError(t, err)

		retrieved, err := GetTowerById(ctx, tower.TowerId)
		assert.NoError(t, err)
		assert.Equal(t, tower.TowerId, retrieved.TowerId)
		assert.Equal(t, tower.Name, retrieved.Name)
		assert.Equal(t, tower.Role, retrieved.Role)
	})

	t.Run("GetNonexistentTower", func(t *testing.T) {
		_, err := GetTowerById(ctx, "nonexistent")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrTowerNotFound)
	})

	t.Run("GetLocal", func(t *testing.T) {
		tower := &Instance{
			TowerId:     primitive.NewObjectID().Hex(),
			Name:        testTowerName,
			Role:        RoleCore,
			IsThisTower: true,
		}

		err := SaveTower(ctx, tower)
		require.NoError(t, err)

		local, err := GetLocal(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, local)
		assert.True(t, local.IsThisTower)
		assert.Equal(t, tower.TowerId, local.TowerId)
	})

	t.Run("GetAllTowersByTowerId", func(t *testing.T) {
		creatorId := primitive.NewObjectID().Hex()
		numTowers := 3

		// Create multiple towers with same creator
		for i := 0; i < numTowers; i++ {
			tower := &Instance{
				TowerId:   primitive.NewObjectID().Hex(),
				Name:      testTowerName,
				Role:      RoleCore,
				CreatedBy: creatorId,
			}
			err := SaveTower(ctx, tower)
			require.NoError(t, err)
		}

		// Create a tower with different creator
		otherTower := &Instance{
			TowerId:   primitive.NewObjectID().Hex(),
			Name:      testTowerName,
			Role:      RoleCore,
			CreatedBy: "different_creator",
		}
		err := SaveTower(ctx, otherTower)
		require.NoError(t, err)

		// Test retrieval
		towers, err := GetAllTowersByTowerId(ctx, creatorId)
		assert.NoError(t, err)
		assert.Len(t, towers, numTowers)

		for _, tower := range towers {
			assert.Equal(t, creatorId, tower.CreatedBy)
		}
	})
}

func TestTower_Updates(t *testing.T) {
	ctx := db.SetupTestDB(t, TowerCollectionKey)

	t.Run("SetLastBackup", func(t *testing.T) {
		tower := &Instance{
			TowerId: primitive.NewObjectID().Hex(),
			Name:    testTowerName,
			Role:    RoleCore,
		}

		err := SaveTower(ctx, tower)
		require.NoError(t, err)

		backupTime := time.Now()
		err = SetLastBackup(ctx, tower.TowerId, backupTime)
		assert.NoError(t, err)

		// Verify update
		updated, err := GetTowerById(ctx, tower.TowerId)
		assert.NoError(t, err)
		assert.Equal(t, backupTime.UnixMilli(), updated.LastBackup)
	})

	t.Run("UpdateRole", func(t *testing.T) {
		tower := &Instance{
			TowerId: primitive.NewObjectID().Hex(),
			Name:    testTowerName,
			Role:    RoleInit,
		}

		err := SaveTower(ctx, tower)
		require.NoError(t, err)

		tower.Role = RoleCore
		err = SaveTower(ctx, tower)
		assert.NoError(t, err)

		// Verify update
		updated, err := GetTowerById(ctx, tower.TowerId)
		assert.NoError(t, err)
		assert.Equal(t, RoleCore, updated.Role)
	})
}

func TestTower_RoleChecks(t *testing.T) {
	ctx := db.SetupTestDB(t, TowerCollectionKey)

	t.Run("RoleValidation", func(t *testing.T) {
		tower := &Instance{
			TowerId: primitive.NewObjectID().Hex(),
			Name:    testTowerName,
			Role:    RoleCore,
		}

		err := SaveTower(ctx, tower)
		require.NoError(t, err)

		assert.True(t, tower.IsCore())
		assert.False(t, tower.IsBackup())

		tower.Role = RoleBackup
		assert.False(t, tower.IsCore())
		assert.True(t, tower.IsBackup())
	})

	t.Run("ReportedRole", func(t *testing.T) {
		tower := &Instance{
			TowerId: primitive.NewObjectID().Hex(),
			Name:    testTowerName,
			Role:    RoleCore,
		}

		tower.SetReportedRole(RoleBackup)
		assert.Equal(t, RoleBackup, tower.GetReportedRole())
	})
}
