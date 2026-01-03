package tower_test

import (
	"context"
	"testing"

	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/structs"
	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/tower"
	"github.com/stretchr/testify/assert"
)

func TestInitializeCoreServer_Validation(t *testing.T) {
	t.Run("returns error when name is empty", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		params := structs.InitServerParams{
			Name:     "",
			Username: "admin",
			Password: "password123",
		}

		err := tower.InitializeCoreServer(appCtx, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required fields")
	})

	t.Run("returns error when username is empty", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		params := structs.InitServerParams{
			Name:     "Test Server",
			Username: "",
			Password: "password123",
		}

		err := tower.InitializeCoreServer(appCtx, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required fields")
	})

	t.Run("returns error when password is empty", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		params := structs.InitServerParams{
			Name:     "Test Server",
			Username: "admin",
			Password: "",
		}

		err := tower.InitializeCoreServer(appCtx, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required fields")
	})

	t.Run("returns error when all fields are empty", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		params := structs.InitServerParams{}

		err := tower.InitializeCoreServer(appCtx, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required fields")
	})
}

func TestTowerRoleConstants(t *testing.T) {
	t.Run("RoleCore is defined", func(t *testing.T) {
		assert.NotEmpty(t, string(tower_model.RoleCore))
	})

	t.Run("RoleBackup is defined", func(t *testing.T) {
		assert.NotEmpty(t, string(tower_model.RoleBackup))
	})

	t.Run("RoleInit is defined", func(_ *testing.T) {
		// RoleInit might be empty string, just ensure it's a valid type
		_ = tower_model.RoleInit
	})
}

func TestTowerErrors(t *testing.T) {
	t.Run("ErrNotCore is defined", func(t *testing.T) {
		assert.NotNil(t, tower_model.ErrNotCore)
	})
}
