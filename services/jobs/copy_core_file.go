package jobs

import (
	"os"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/job"
	task_model "github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/errors"
	task_mod "github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	context_service "github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/notify"
	tower_service "github.com/ethanrous/weblens/services/tower"
	"github.com/rs/zerolog"
)

func CopyFileFromCore(tsk task_mod.Task) {
	t := tsk.(*task_model.Task)
	meta := t.GetMeta().(job.BackupCoreFileMeta)

	ctx, ok := context_service.FromContext(t.Ctx)
	if !ok {
		t.Fail(errors.New("Failed to cast context to FilerContext"))

		return
	}

	t.SetErrorCleanup(func(tsk task_mod.Task) {
		t := tsk.(*task_model.Task)
		failNotif := notify.NewTaskNotification(t, websocket_mod.CopyFileFailedEvent, task_mod.TaskResult{"filename": meta.Filename, "coreId": meta.Core.TowerId})
		ctx.Notify(ctx, failNotif)

		rmErr := ctx.FileService.DeleteFiles(t.Ctx, meta.File)
		if rmErr != nil {
			t.Log().Error().Stack().Err(rmErr).Msg("")
		}
	})

	filename := meta.Filename
	if filename == "" {
		filename = meta.File.GetPortablePath().Filename()
	}

	if meta.File.GetContentId() == "" {
		t.Fail(errors.WithStack(file_model.ErrNoContentId))

		return
	}

	ctx.Notify(ctx,
		notify.NewPoolNotification(
			t.GetTaskPool(),
			websocket_mod.CopyFileStartedEvent,
			task_mod.TaskResult{"filename": filename, "coreId": meta.Core.TowerId, "timestamp": time.Now().UnixMilli()},
		),
	)

	t.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Copying file from core [%s]", meta.File.GetPortablePath().Filename()) })

	restoreFile, err := ctx.FileService.NewBackupRestoreFile(ctx, meta.File.GetContentId(), meta.Core.TowerId)
	if err != nil {
		t.Fail(err)

		return
	}

	writeFile, err := restoreFile.Writer()
	if err != nil {
		t.Fail(err)

		return
	}

	defer writeFile.Close()

	err = tower_service.DownloadFileFromCore(ctx, meta.Core, meta.CoreFileId, writeFile)
	if err != nil {
		t.Fail(err)

		return
	}

	err = os.Link(restoreFile.GetPortablePath().ToAbsolute(), meta.File.GetPortablePath().ToAbsolute())
	if err != nil {
		t.Fail(err)

		return
	}

	poolProgress := getScanResult(t)
	poolProgress["filename"] = filename
	poolProgress["coreId"] = meta.Core.TowerId

	notif := notify.NewPoolNotification(t.GetTaskPool(), websocket_mod.CopyFileCompleteEvent, poolProgress)
	if notif.SubscribeKey == "" {
		ctx.Log().Error().Msg("Failed to get subscribe key for pool notification")
	}

	ctx.Notify(ctx, notif)

	t.Success()
}
