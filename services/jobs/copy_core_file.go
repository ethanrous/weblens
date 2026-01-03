package jobs

import (
	"os"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	websocket_mod "github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/notify"
	tower_service "github.com/ethanrous/weblens/services/tower"
	"github.com/rs/zerolog"
)

// CopyFileFromCore downloads a file from a core server during a backup operation.
func CopyFileFromCore(tsk *task.Task) {
	meta := tsk.GetMeta().(job.BackupCoreFileMeta)

	ctx, ok := context_service.FromContext(tsk.Ctx)
	if !ok {
		tsk.Fail(wlerrors.New("Failed to cast context to FilerContext"))

		return
	}

	tsk.SetErrorCleanup(func(tsk *task.Task) {
		failNotif := notify.NewTaskNotification(tsk, websocket_mod.CopyFileFailedEvent, task.Result{"filename": meta.Filename, "coreID": meta.Core.TowerID})
		ctx.Notify(ctx, failNotif)

		rmErr := ctx.FileService.DeleteFiles(tsk.Ctx, meta.File)
		if rmErr != nil {
			tsk.Log().Error().Stack().Err(rmErr).Msg("")
		}
	})

	filename := meta.Filename
	if filename == "" {
		filename = meta.File.GetPortablePath().Filename()
	}

	if meta.File.GetContentID() == "" {
		tsk.Fail(wlerrors.WithStack(file_model.ErrNoContentID))

		return
	}

	ctx.Notify(ctx,
		notify.NewPoolNotification(
			tsk.GetTaskPool(),
			websocket_mod.CopyFileStartedEvent,
			task.Result{"filename": filename, "coreID": meta.Core.TowerID, "timestamp": time.Now().UnixMilli()},
		),
	)

	tsk.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Copying file from core [%s]", meta.File.GetPortablePath().Filename()) })

	restoreFile, err := ctx.FileService.NewBackupRestoreFile(ctx, meta.File.GetContentID(), meta.Core.TowerID)
	if err != nil {
		tsk.Fail(err)

		return
	}

	writeFile, err := restoreFile.Writer()
	if err != nil {
		tsk.Fail(err)

		return
	}

	defer writeFile.Close() //nolint:errcheck

	err = tower_service.DownloadFileFromCore(ctx, meta.Core, meta.CoreFileID, writeFile)
	if err != nil {
		tsk.Fail(err)

		return
	}

	err = os.Link(restoreFile.GetPortablePath().ToAbsolute(), meta.File.GetPortablePath().ToAbsolute())
	if err != nil {
		tsk.Fail(err)

		return
	}

	poolProgress := getScanResult(tsk)
	poolProgress["filename"] = filename
	poolProgress["coreID"] = meta.Core.TowerID

	notif := notify.NewPoolNotification(tsk.GetTaskPool(), websocket_mod.CopyFileCompleteEvent, poolProgress)
	if notif.SubscribeKey == "" {
		ctx.Log().Error().Msg("Failed to get subscribe key for pool notification")
	}

	ctx.Notify(ctx, notif)

	tsk.Success()
}
