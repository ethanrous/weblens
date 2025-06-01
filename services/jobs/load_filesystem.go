package jobs

import (
	job_model "github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/errors"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/ethanrous/weblens/services/context"
)

func LoadAtPath(tsk task_mod.Task) {
	t := tsk.(*task.Task)

	appCtx, ok := context.FromContext(t.Ctx)
	if !ok {
		t.Fail(errors.Errorf("failed to get context"))

		return
	}

	meta, ok := t.GetMeta().(job_model.LoadFilesystemMeta)
	if !ok {
		t.Fail(errors.Errorf("failed to get meta"))

		return
	}

	appCtx.Log().Debug().Msgf("Loading filesystem at path %s", meta.File.GetPortablePath())

	_, err := appCtx.FileService.GetChildren(appCtx, meta.File)
	if err != nil {
		t.Fail(err)

		return
	}

	t.Success()
}
