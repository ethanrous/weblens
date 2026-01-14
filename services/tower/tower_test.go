package tower_test

import (
	"context"
	"testing"

	"github.com/ethanrous/weblens/modules/config"
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

		err := tower.InitializeCoreServer(appCtx, params, config.Provider{})
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

		err := tower.InitializeCoreServer(appCtx, params, config.Provider{})
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

		err := tower.InitializeCoreServer(appCtx, params, config.Provider{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required fields")
	})

	t.Run("returns error when all fields are empty", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		params := structs.InitServerParams{}

		err := tower.InitializeCoreServer(appCtx, params, config.Provider{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required fields")
	})
}
