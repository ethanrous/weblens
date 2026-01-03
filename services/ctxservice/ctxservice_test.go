package ctxservice_test

import (
	"context"
	"testing"

	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBasicContext(t *testing.T) {
	t.Run("creates basic context with logger", func(t *testing.T) {
		logger := log.NewZeroLogger()
		ctx := ctxservice.NewBasicContext(context.Background(), logger)

		assert.NotNil(t, ctx)
		assert.NotNil(t, ctx.Log())
	})
}

func TestBasicContext_Log(t *testing.T) {
	t.Run("returns logger from context", func(t *testing.T) {
		logger := log.NewZeroLogger()
		ctx := ctxservice.NewBasicContext(context.Background(), logger)

		retrieved := ctx.Log()
		assert.NotNil(t, retrieved)
	})
}

func TestBasicContext_WithValue(t *testing.T) {
	t.Run("adds value to context", func(t *testing.T) {
		logger := log.NewZeroLogger()
		ctx := ctxservice.NewBasicContext(context.Background(), logger)

		key := "testKey"
		value := "testValue"
		newCtx := ctx.WithValue(key, value)

		assert.Equal(t, value, newCtx.Value(key))
	})

	t.Run("preserves original context", func(t *testing.T) {
		logger := log.NewZeroLogger()
		ctx := ctxservice.NewBasicContext(context.Background(), logger)

		key := "testKey"
		value := "testValue"
		_ = ctx.WithValue(key, value)

		// Original context should not have the value
		assert.Nil(t, ctx.Value(key))
	})
}

func TestNewAppContext(t *testing.T) {
	t.Run("creates app context from basic context", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)

		appCtx := ctxservice.NewAppContext(basicCtx)

		assert.NotNil(t, appCtx)
		assert.NotNil(t, appCtx.Log())
	})
}

func TestFromContext(t *testing.T) {
	t.Run("extracts app context", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		// Use the app context as a regular context
		extracted, ok := ctxservice.FromContext(appCtx)

		assert.True(t, ok)
		assert.NotNil(t, extracted)
	})

	t.Run("returns false for non-app context", func(t *testing.T) {
		ctx := context.Background()

		_, ok := ctxservice.FromContext(ctx)

		assert.False(t, ok)
	})

	t.Run("returns false for nil context", func(t *testing.T) {
		_, ok := ctxservice.FromContext(nil)

		assert.False(t, ok)
	})
}

func TestAppContext_WithValue(t *testing.T) {
	t.Run("adds value to app context", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		key := "appKey"
		value := "appValue"
		newAppCtx := appCtx.WithValue(key, value)

		assert.Equal(t, value, newAppCtx.Value(key))
	})
}

func TestAppContext_Database(t *testing.T) {
	t.Run("returns nil database when not set", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		db := appCtx.Database()
		assert.Nil(t, db)
	})
}

func TestAppContext_GetCache(t *testing.T) {
	t.Run("creates cache for collection", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		cache := appCtx.GetCache("testCollection")
		assert.NotNil(t, cache)
	})

	t.Run("returns same cache for same collection", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		cache1 := appCtx.GetCache("testCollection")
		cache2 := appCtx.GetCache("testCollection")

		assert.Same(t, cache1, cache2)
	})

	t.Run("returns different caches for different collections", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		cache1 := appCtx.GetCache("collection1")
		cache2 := appCtx.GetCache("collection2")

		assert.NotSame(t, cache1, cache2)
	})
}

func TestAppContext_ClearCache(t *testing.T) {
	t.Run("clears all caches", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		// Create some caches
		appCtx.GetCache("collection1")
		appCtx.GetCache("collection2")

		// Clear all caches
		appCtx.ClearCache()

		// New caches should be created
		newCache1 := appCtx.GetCache("collection1")
		assert.NotNil(t, newCache1)
	})
}

func TestAppContext_GetTowerID(t *testing.T) {
	t.Run("returns empty tower ID when not set", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		towerID := appCtx.GetTowerID()
		assert.Empty(t, towerID)
	})
}

func TestAppContext_GetFileService(t *testing.T) {
	t.Run("returns nil file service when not set", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		fs := appCtx.GetFileService()
		assert.Nil(t, fs)
	})
}

func TestAppContext_WithContext(t *testing.T) {
	t.Run("combines app context with new context", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		newBaseCtx := context.WithValue(context.Background(), "newKey", "newValue")
		combined := appCtx.WithContext(newBaseCtx)

		// The combined context should be usable
		assert.NotNil(t, combined)
	})
}

func TestAppContext_GetMongoSession(t *testing.T) {
	t.Run("returns nil session when not set", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		session := appCtx.GetMongoSession()
		assert.Nil(t, session)
	})
}

func TestErrNoContext(t *testing.T) {
	t.Run("error is defined", func(t *testing.T) {
		assert.NotNil(t, ctxservice.ErrNoContext)
		assert.Contains(t, ctxservice.ErrNoContext.Error(), "not an AppContext")
	})
}

func TestAppContext_LoggerIntegration(t *testing.T) {
	t.Run("logger can be used for logging", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		// This should not panic
		require.NotPanics(t, func() {
			appCtx.Log().Debug().Msg("test message")
		})
	})
}

func TestAppContext_Value(t *testing.T) {
	t.Run("returns app context for appContextKey", func(t *testing.T) {
		logger := log.NewZeroLogger()
		basicCtx := ctxservice.NewBasicContext(context.Background(), logger)
		appCtx := ctxservice.NewAppContext(basicCtx)

		// The context should return itself for the appContextKey
		extracted, ok := ctxservice.FromContext(appCtx)
		assert.True(t, ok)
		assert.NotNil(t, extracted)
	})
}

func TestNewZeroLogger(t *testing.T) {
	t.Run("creates valid zerolog logger", func(t *testing.T) {
		logger := log.NewZeroLogger()
		assert.IsType(t, &zerolog.Logger{}, logger)
	})
}
