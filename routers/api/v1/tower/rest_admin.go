package tower

import (
	"net/http"

	"github.com/ethanrous/weblens/models/config"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/net"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	"github.com/ethanrous/weblens/modules/structs"

	"github.com/ethanrous/weblens/services/context"
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
func GetRunningTasks(ctx context.RequestContext) {
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
func FlushCache(ctx context.RequestContext) {
	ctx.ClearCache()
	ctx.JSON(http.StatusOK, structs.WLResponseInfo{Message: "Cache flushed successfully"})
}

// GetConfig godoc
//
//	@ID			GetConfig
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Summary	Get Config
//	@Tags		Config
//	@Produce	json
//
//	@Success	200 {object}	config.Config "Config Info"
//	@Router		/config [get]
func GetConfig(ctx context.RequestContext) {
	cnf, err := config.GetConfig(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.Errorf("Failed to get config: %w", err))

		return
	}

	ctx.JSON(http.StatusOK, cnf)
}

// SetConfig godoc
//
//	@ID			SetConfig
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Param		request	body	structs.SetConfigParams	true	"Set Config Params"
//
//	@Summary	Set Config
//	@Tags		Config
//	@Produce	json
//
//	@Success	200
//	@Router		/config [post]
func SetConfig(ctx context.RequestContext) {
	configParams, err := net.ReadRequestBody[structs.SetConfigParams](ctx.Req)
	if err != nil {
		return
	}

	cnf, err := config.GetConfig(ctx)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.Errorf("Failed to get config: %w", err))

		return
	}

	for _, param := range configParams {
		switch param.ConfigKey {
		case config.AllowRegistrations:
			cnf.AllowRegistrations = param.ConfigValue.(bool)
		case config.EnableHDIR:
			cnf.EnableHDIR = param.ConfigValue.(bool)
		default:
			ctx.Error(http.StatusBadRequest, errors.Errorf("Unknown config parameter: %s", param.ConfigKey))
		}
	}

	err = config.SaveConfig(ctx, cnf)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.Errorf("Failed to save config: %w", err))

		return
	}
}
