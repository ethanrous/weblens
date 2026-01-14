package tower_test

import (
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/ethanrous/weblens/models/tower"
)

const (
	testTowerName = "Test Tower"
)

func TestTower_Creation(t *testing.T) {
	ctx := db.SetupTestDB(t, tower.TowerCollectionKey)

	t.Run("CreateLocal", func(t *testing.T) {
		instance, err := tower.CreateLocal(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, instance)
		assert.Equal(t, tower.RoleUninitialized, instance.Role)
		assert.True(t, instance.IsThisTower)

		// Verify tower was saved to database
		savedTower, err := tower.GetTowerByID(ctx, instance.TowerID)
		assert.NoError(t, err)
		assert.Equal(t, instance.TowerID, savedTower.TowerID)
	})

	t.Run("CreateTower", func(t *testing.T) {
		instance := &tower.Instance{
			TowerID: primitive.NewObjectID().Hex(),
			Name:    testTowerName,
			Role:    tower.RoleCore,
		}

		err := tower.SaveTower(ctx, instance)
		assert.NoError(t, err)

		// Verify tower was saved
		savedTower, err := tower.GetTowerByID(ctx, instance.TowerID)
		assert.NoError(t, err)
		assert.Equal(t, instance.TowerID, savedTower.TowerID)
		assert.Equal(t, testTowerName, savedTower.Name)
	})

	t.Run("CreateInvalidTower", func(t *testing.T) {
		instance := &tower.Instance{
			TowerID: primitive.NewObjectID().Hex(),
			Role:    tower.RoleCore,
			// Missing Name
		}

		err := tower.SaveTower(ctx, instance)
		assert.Error(t, err)
	})
}

func TestTower_Retrieval(t *testing.T) {
	ctx := db.SetupTestDB(t, tower.TowerCollectionKey)

	t.Run("GetTowerByID", func(t *testing.T) {
		instance := &tower.Instance{
			TowerID: primitive.NewObjectID().Hex(),
			Name:    testTowerName,
			Role:    tower.RoleCore,
		}

		err := tower.SaveTower(ctx, instance)
		require.NoError(t, err)

		retrieved, err := tower.GetTowerByID(ctx, instance.TowerID)
		assert.NoError(t, err)
		assert.Equal(t, instance.TowerID, retrieved.TowerID)
		assert.Equal(t, instance.Name, retrieved.Name)
		assert.Equal(t, instance.Role, retrieved.Role)
	})

	t.Run("GetNonexistentTower", func(t *testing.T) {
		_, err := tower.GetTowerByID(ctx, "nonexistent")
		assert.Error(t, err)
		assert.ErrorIs(t, err, tower.ErrTowerNotFound)
	})

	t.Run("GetLocal", func(t *testing.T) {
		instance := &tower.Instance{
			TowerID:     primitive.NewObjectID().Hex(),
			Name:        testTowerName,
			Role:        tower.RoleCore,
			IsThisTower: true,
		}

		err := tower.SaveTower(ctx, instance)
		require.NoError(t, err)

		local, err := tower.GetLocal(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, local)
		assert.True(t, local.IsThisTower)
		assert.Equal(t, instance.TowerID, local.TowerID)
	})

	t.Run("GetAllTowersByTowerID", func(t *testing.T) {
		creatorID := primitive.NewObjectID().Hex()
		numTowers := 3

		// Create multiple towers with same creator
		for range numTowers {
			instance := &tower.Instance{
				TowerID:   primitive.NewObjectID().Hex(),
				Name:      testTowerName,
				Role:      tower.RoleCore,
				CreatedBy: creatorID,
			}
			err := tower.SaveTower(ctx, instance)
			require.NoError(t, err)
		}

		// Create a tower with different creator
		otherTower := &tower.Instance{
			TowerID:   primitive.NewObjectID().Hex(),
			Name:      testTowerName,
			Role:      tower.RoleCore,
			CreatedBy: "different_creator",
		}
		err := tower.SaveTower(ctx, otherTower)
		require.NoError(t, err)

		// Test retrieval
		towers, err := tower.GetAllTowersByTowerID(ctx, creatorID)
		assert.NoError(t, err)
		assert.Len(t, towers, numTowers)

		for _, tower := range towers {
			assert.Equal(t, creatorID, tower.CreatedBy)
		}
	})
}

func TestTower_Updates(t *testing.T) {
	ctx := db.SetupTestDB(t, tower.TowerCollectionKey)

	t.Run("SetLastBackup", func(t *testing.T) {
		instance := &tower.Instance{
			TowerID: primitive.NewObjectID().Hex(),
			Name:    testTowerName,
			Role:    tower.RoleCore,
		}

		err := tower.SaveTower(ctx, instance)
		require.NoError(t, err)

		backupTime := time.Now()
		err = tower.SetLastBackup(ctx, instance.TowerID, backupTime, 10)
		assert.NoError(t, err)

		// Verify update
		updated, err := tower.GetTowerByID(ctx, instance.TowerID)
		assert.NoError(t, err)
		assert.Equal(t, backupTime.UnixMilli(), updated.LastBackup)
	})

	t.Run("UpdateRole", func(t *testing.T) {
		instance := &tower.Instance{
			TowerID: primitive.NewObjectID().Hex(),
			Name:    testTowerName,
			Role:    tower.RoleUninitialized,
		}

		err := tower.SaveTower(ctx, instance)
		require.NoError(t, err)

		instance.Role = tower.RoleCore
		err = tower.SaveTower(ctx, instance)
		assert.NoError(t, err)

		// Verify update
		updated, err := tower.GetTowerByID(ctx, instance.TowerID)
		assert.NoError(t, err)
		assert.Equal(t, tower.RoleCore, updated.Role)
	})
}

func TestTower_RoleChecks(t *testing.T) {
	ctx := db.SetupTestDB(t, tower.TowerCollectionKey)

	t.Run("RoleValidation", func(t *testing.T) {
		instance := &tower.Instance{
			TowerID: primitive.NewObjectID().Hex(),
			Name:    testTowerName,
			Role:    tower.RoleCore,
		}

		err := tower.SaveTower(ctx, instance)
		require.NoError(t, err)

		assert.True(t, instance.IsCore())
		assert.False(t, instance.IsBackup())

		instance.Role = tower.RoleBackup
		assert.False(t, instance.IsCore())
		assert.True(t, instance.IsBackup())
	})

	t.Run("ReportedRole", func(t *testing.T) {
		instance := &tower.Instance{
			TowerID: primitive.NewObjectID().Hex(),
			Name:    testTowerName,
			Role:    tower.RoleCore,
		}

		instance.SetReportedRole(tower.RoleBackup)
		assert.Equal(t, tower.RoleBackup, instance.GetReportedRole())
	})
}
