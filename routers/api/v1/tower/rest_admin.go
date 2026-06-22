// Package tower contains admin REST API handlers for tower management.
package tower

import (
	"net/http"
	"strconv"

	"github.com/ethanrous/weblens/models/featureflags"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/netwrk"
	"github.com/ethanrous/weblens/modules/wlerrors"
	slices_mod "github.com/ethanrous/weblens/modules/wlslices"
	"github.com/ethanrous/weblens/modules/wlstructs"

	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/reshape"
)

// FilterTasks returns the tasks a poller should see: every still-active task, plus —
// when includeExited is set — finished tasks. A positive sinceMs further limits finished
// tasks to those that completed at or after that Unix epoch-ms cursor, so the gantt can poll
// incrementally instead of re-fetching the whole retained history every time. The bound is
// inclusive (>=) and the client dedups by task ID, so a task finishing in the same millisecond
// as the cursor is re-returned rather than lost.
func FilterTasks(tasks []*task.Task, includeExited bool, sinceMs int64) []*task.Task {
	return slices_mod.Filter(tasks, func(t *task.Task) bool {
		if t.QueueState() != task.Exited {
			return true
		}

		if !includeExited {
			return false
		}

		return sinceMs <= 0 || t.GetFinishTime().UnixMilli() >= sinceMs
	})
}

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
//	@Param		includeExited	query	bool	false	"Include tasks that have already finished (still held in memory)"					default(false)
//	@Param		since			query	int64	false	"Only return finished tasks that completed at or after this Unix epoch-ms cursor (incremental polling)"	default(0)
//
//	@Success	200	{array}	wlstructs.TaskInfo	"Task Infos"
//	@Router		/tower/tasks [get]
func GetRunningTasks(ctx ctxservice.RequestContext) {
	var sinceMs int64

	if s := ctx.Query("since"); s != "" {
		parsed, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			ctx.Error(http.StatusBadRequest, wlerrors.New("invalid since format"))

			return
		}

		sinceMs = parsed
	}

	tasks := FilterTasks(ctx.TaskService.GetTasks(), ctx.QueryBool("includeExited"), sinceMs)

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
//	@Success	200	{object}	wlstructs.WLResponseInfo	"Cache flushed successfully"
//	@Router		/tower/cache [delete]
func FlushCache(ctx ctxservice.RequestContext) {
	ctx.ClearCache()
	ctx.JSON(http.StatusOK, wlstructs.WLResponseInfo{Message: "Cache flushed successfully"})
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
//	@Success	200	{object}	featureflags.Bundle	"Feature Flags"
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
//	@Param		request	body	wlstructs.FeatureFlagParams	true	"Feature Flag Params"
//
//	@Summary	Set Feature Flags
//	@Tags		FeatureFlags
//	@Produce	json
//
//	@Success	200
//	@Router		/flags [post]
func SetFlags(ctx ctxservice.RequestContext) {
	configParams, err := netwrk.ReadRequestBody[wlstructs.FeatureFlagParams](ctx.Req)
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
		case featureflags.EnableEmbed:
			cnf.EnableEmbed = param.ConfigValue.(bool)
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
