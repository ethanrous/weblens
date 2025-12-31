package jobs

import (
	"os"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/job"
	task_model "github.com/ethanrous/weblens/models/task"
	task_mod "github.com/ethanrous/weblens/modules/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/notify"
	tower_service "github.com/ethanrous/weblens/services/tower"
	"github.com/rs/zerolog"
)

// CopyFileFromCore downloads a file from a core server during a backup operation.
func CopyFileFromCore(tsk task_mod.Task) {
	t := tsk.(*task_model.Task)
	meta := t.GetMeta().(job.BackupCoreFileMeta)

	ctx, ok := context_service.FromContext(t.Ctx)
	if !ok {
		t.Fail(wlerrors.New("Failed to cast context to FilerContext"))

		return
	}

	t.SetErrorCleanup(func(tsk task_mod.Task) {
		t := tsk.(*task_model.Task)
		failNotif := notify.NewTaskNotification(t, websocket_mod.CopyFileFailedEvent, task_mod.Result{"filename": meta.Filename, "coreID": meta.Core.TowerID})
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

	if meta.File.GetContentID() == "" {
		t.Fail(wlerrors.WithStack(file_model.ErrNoContentID))

		return
	}

	ctx.Notify(ctx,
		notify.NewPoolNotification(
			t.GetTaskPool(),
			websocket_mod.CopyFileStartedEvent,
			task_mod.Result{"filename": filename, "coreID": meta.Core.TowerID, "timestamp": time.Now().UnixMilli()},
		),
	)

	t.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Copying file from core [%s]", meta.File.GetPortablePath().Filename()) })

	restoreFile, err := ctx.FileService.NewBackupRestoreFile(ctx, meta.File.GetContentID(), meta.Core.TowerID)
	if err != nil {
		t.Fail(err)

		return
	}

	writeFile, err := restoreFile.Writer()
	if err != nil {
		t.Fail(err)

		return
	}

	defer writeFile.Close() //nolint:errcheck

	err = tower_service.DownloadFileFromCore(ctx, meta.Core, meta.CoreFileID, writeFile)
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
	poolProgress["coreID"] = meta.Core.TowerID

	notif := notify.NewPoolNotification(t.GetTaskPool(), websocket_mod.CopyFileCompleteEvent, poolProgress)
	if notif.SubscribeKey == "" {
		ctx.Log().Error().Msg("Failed to get subscribe key for pool notification")
	}

	ctx.Notify(ctx, notif)

	t.Success()
}
