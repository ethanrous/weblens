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
	exists := wlerrors.Is(err, file_model.ErrFileAlreadyExists)

	if err != nil && !exists {
		tsk.Fail(err)

		return
	}

	var bytesCopied int64

	if err != nil {
		ctx.Log().Trace().Msgf("Restore file already exists at [%s], not copying", restoreFile.GetPortablePath().ToAbsolute())
	} else {
		writeFile, err := restoreFile.Writer()
		if err != nil {
			tsk.Fail(err)

			return
		}

		defer writeFile.Close() //nolint:errcheck

		bytesCopied, err = tower_service.DownloadFileFromCore(ctx, meta.Core, meta.CoreFileID, writeFile)
		if err != nil {
			tsk.Fail(err)

			return
		}
	}

	tsk.Log().Trace().Func(func(e *zerolog.Event) {
		e.Msgf("Copy file is linking %s -> %s", restoreFile.GetPortablePath().ToAbsolute(), meta.File.GetPortablePath().ToAbsolute())
	})

	err = os.Link(restoreFile.GetPortablePath().ToAbsolute(), meta.File.GetPortablePath().ToAbsolute())
	if err != nil {
		tsk.Fail(err)

		return
	}

	if !exists {
		err = ctx.FileService.AddFile(tsk.Ctx, meta.File)
		if err != nil {
			tsk.Fail(err)

			return
		}
	}

	tsk.SetResult(task.Result{
		"bytesCopied": bytesCopied,
	})
	tsk.Success()
}
