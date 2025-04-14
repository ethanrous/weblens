package jobs

import (
	"io"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/job"
	task_model "github.com/ethanrous/weblens/models/task"
	task_mod "github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/ethanrous/weblens/services/proxy"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func CopyFileFromCore(tsk task_mod.Task) {
	t := tsk.(*task_model.Task)
	meta := t.GetMeta().(job.BackupCoreFileMeta)

	ctx, ok := t.Ctx.(*context.AppContext)
	if !ok {
		t.Fail(errors.New("Failed to cast context to FilerContext"))
		return
	}

	t.SetErrorCleanup(func(tsk task_mod.Task) {
		t := tsk.(*task_model.Task)
		failNotif := notify.NewTaskNotification(t, websocket_mod.CopyFileFailedEvent, task_mod.TaskResult{"filename": meta.Filename, "coreId": meta.Core.TowerId})
		t.Ctx.Notify(failNotif)

		rmErr := ctx.FileService.DeleteFiles(t.Ctx, []*file_model.WeblensFileImpl{meta.File})
		if rmErr != nil {
			t.Ctx.Log().Error().Stack().Err(rmErr).Msg("")
		}
	})

	filename := meta.Filename
	if filename == "" {
		filename = meta.File.GetPortablePath().Filename()
	}

	t.Ctx.Notify(
		notify.NewPoolNotification(
			t.GetTaskPool(),
			websocket_mod.CopyFileStartedEvent,
			task_mod.TaskResult{"filename": filename, "coreId": meta.Core.TowerId, "timestamp": time.Now().UnixMilli()},
		),
	)

	t.Ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Copying file from core [%s]", meta.File.GetPortablePath().Filename()) })

	if meta.File.GetContentId() == "" {
		t.ReqNoErr(errors.WithStack(file_model.ErrNoContentId))
	}

	writeFile, err := meta.File.Writeable()
	if err != nil {
		t.Fail(err)
	}
	defer writeFile.Close()

	res, err := proxy.NewCoreRequest(meta.Core, "GET", "/files/"+meta.CoreFileId+"/download").Call()
	if err != nil {
		t.Fail(err)
	}

	defer res.Body.Close()

	_, err = io.Copy(writeFile, res.Body)
	if err != nil {
		t.Fail(err)
	}

	poolProgress := getScanResult(t)
	poolProgress["filename"] = filename
	poolProgress["coreId"] = meta.Core.TowerId

	t.Ctx.Notify(notify.NewPoolNotification(t.GetTaskPool(), websocket_mod.CopyFileCompleteEvent, poolProgress))

	t.Success()
}
