package jobs

import (
	job_model "github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/services/ctxservice"
	file_service "github.com/ethanrous/weblens/services/file"
)

// LoadAtPath recursively loads the filesystem tree starting from a specified path.
func LoadAtPath(tsk task_mod.Task) {
	t := tsk.(*task.Task)

	appCtx, ok := ctxservice.FromContext(t.Ctx)
	if !ok {
		t.Fail(wlerrors.Errorf("failed to get context"))

		return
	}

	meta, ok := t.GetMeta().(job_model.LoadFilesystemMeta)
	if !ok {
		t.Fail(wlerrors.Errorf("failed to get meta"))

		return
	}

	appCtx.Log().Debug().Msgf("Loading filesystem at path %s", meta.File.GetPortablePath())

	err := file_service.LoadFilesRecursively(appCtx, meta.File)
	if err != nil {
		t.Fail(err)

		return
	}

	t.Success()
}
