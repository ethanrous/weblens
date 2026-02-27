// Package tower contains admin REST API handlers for tower management.
package tower

import (
	"net/http"

	"github.com/ethanrous/weblens/models/featureflags"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/netwrk"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	"github.com/ethanrous/weblens/modules/structs"
	"github.com/ethanrous/weblens/modules/wlerrors"

	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/reshape"
)

// GetRunningTasks godoc
//
//	@ID			GetRunningTasks
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Summary	Get Running Tasks
//	@Tags		Towers
//	@Produce	json
//
//	@Success	200		{array}	structs.TaskInfo	"Task Infos"
//	@Router		/tower/tasks [get]
func GetRunningTasks(ctx ctxservice.RequestContext) {
	tasksIter := ctx.TaskService.GetTasks()

	tasks := slices_mod.Filter(tasksIter, func(t *task.Task) bool {
		return t.QueueState() != task.Exited
	})

	taskInfos := reshape.TasksToTaskInfos(tasks)
	ctx.JSON(http.StatusOK, taskInfos)
}

// FlushCache godoc
//
//	@ID			FlushCache
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Summary	Flush Cache
//	@Tags		Towers
//	@Produce	json
//
//	@Success	200		{object}	structs.WLResponseInfo	"Cache flushed successfully"
//	@Router		/tower/cache [delete]
func FlushCache(ctx ctxservice.RequestContext) {
	ctx.ClearCache()
	ctx.JSON(http.StatusOK, structs.WLResponseInfo{Message: "Cache flushed successfully"})
}

// GetFlags godoc
//
//	@ID			GetFlags
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Summary	Get Feature Flags
//	@Tags		FeatureFlags
//	@Produce	json
//
//	@Success	200 {object}	featureflags.Bundle "Feature Flags"
//	@Router		/flags [get]
func GetFlags(ctx ctxservice.RequestContext) {
	cnf, err := featureflags.GetFlags(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.Errorf("Failed to get config: %w", err))

		return
	}

	ctx.JSON(http.StatusOK, cnf)
}

// SetFlags godoc
//
//	@ID			SetFlags
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Param		request	body	structs.FeatureFlagParams	true	"Feature Flag Params"
//
//	@Summary	Set Feature Flags
//	@Tags		FeatureFlags
//	@Produce	json
//
//	@Success	200
//	@Router		/flags [post]
func SetFlags(ctx ctxservice.RequestContext) {
	configParams, err := netwrk.ReadRequestBody[structs.FeatureFlagParams](ctx.Req)
	if err != nil {
		return
	}

	cnf, err := featureflags.GetFlags(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.Errorf("Failed to get feature flags: %w", err))

		return
	}

	for _, param := range configParams {
		switch param.ConfigKey {
		case featureflags.AllowRegistrations:
			cnf.AllowRegistrations = param.ConfigValue.(bool)
		case featureflags.EnableHDIR:
			cnf.EnableHDIR = param.ConfigValue.(bool)
		case featureflags.EnableWebDAV:
			cnf.EnableWebDAV = param.ConfigValue.(bool)
		default:
			ctx.Error(http.StatusBadRequest, wlerrors.Errorf("Unknown feature flag: %s", param.ConfigKey))
		}
	}

	err = featureflags.SaveFlags(ctx, cnf)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.Errorf("Failed to save feature flags: %w", err))

		return
	}
}
