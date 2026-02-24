package jobs

import (
	"slices"
	"time"

	"github.com/ethanrous/weblens/models/featureflags"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/job"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	media_service "github.com/ethanrous/weblens/services/media"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ScanDirectory scans a directory and processes all files within it.
func ScanDirectory(t *task.Task) {
	ctx, ok := context_service.FromContext(t.Ctx)
	if !ok {
		t.Fail(wlerrors.New("failed to get context"))

		return
	}

	meta := t.GetMeta().(job.ScanMeta)

	t.SetErrorCleanup(func(_ *task.Task) {
		err := t.ReadError()
		notif := notify.NewTaskNotification(t, websocket.TaskFailedEvent, task.Result{"error": err.Error()})
		ctx.Notify(t.Ctx, notif)
	})

	if file_model.IsFileInTrash(meta.File) {
		// Let any client subscribers know we are done
		notif := notify.NewTaskNotification(t, websocket.FolderScanCompleteEvent, task.Result{"executionTime": t.ExeTime()})
		ctx.Notify(ctx, notif)
		t.Success("No media to scan")

		return
	}

	// Create a new task pool for the file scans this directory scan will spawn
	pool, err := t.GetTaskPool().GetWorkerPool().NewTaskPool(true, t)
	if err != nil {
		t.Fail(wlerrors.WithStack(err))

		return
	}

	ctx.ClientService.FolderSubToTask(ctx, meta.File.ID(), t)

	parent := meta.File.GetParent()
	if parent != nil {
		ctx.ClientService.FolderSubToTask(ctx, parent.ID(), t)
	}

	t.SetCleanup(func(tsk *task.Task) {
		// Make sure we finish sending any messages to the client
		// before we close unsubscribe from the task
		ctx.ClientService.Flush(t.Ctx)

		err := ctx.ClientService.UnsubscribeAllByID(t.Ctx, tsk.ID(), websocket.TaskSubscribe)
		if err != nil {
			tsk.Log().Error().Err(err).Msg("Failed to unsubscribe from task")
		}
	})

	t.Log().Debug().Func(func(e *zerolog.Event) {
		e.Msgf("Beginning directory scan for [%s] (%s)", meta.File.GetPortablePath(), meta.File.ID())
	})

	var alreadyFiles []*file_model.WeblensFileImpl

	var alreadyMedia []*media_model.Media

	start := time.Now()

	cnf, err := featureflags.GetFlags(ctx)
	if err != nil {
		t.Fail(wlerrors.WithStack(err))

		return
	}

	t.SetResult(task.Result{
		"filepath": meta.File.GetPortablePath().String(),
		"state":    "Discovering files",
	})

	err = meta.File.LeafMap(
		ctx,
		ctx.FileService,
		func(mf *file_model.WeblensFileImpl) error {
			return queueScanFileIfNeeded(ctx, t, mf, cnf.EnableHDIR, &alreadyFiles, &alreadyMedia, pool)
		},
	)

	t.SetResult(task.Result{
		"filepath": meta.File.GetPortablePath().String(),
		"state":    "Done discovering files",
	})

	log.Debug().Func(func(e *zerolog.Event) {
		e.Str("portable_file_path", meta.File.GetPortablePath().String()).Msgf("Directory scan found files in %s", time.Since(start))
	})

	// If the files are already in the media service, we need to update the clients
	// that may be waiting for these files to be processed, but since we won't be
	// adding those via the media service, we need to update the clients now
	if len(alreadyFiles) > 0 {
		updates := make([]websocket.WsResponseInfo, 0, len(alreadyFiles))

		for i := range alreadyFiles {
			mediaInfo := reshape.MediaToMediaInfo(alreadyMedia[i])
			o := notify.FileNotificationOptions{MediaInfo: mediaInfo}

			fInfo, err := reshape.WeblensFileToFileInfo(ctx, alreadyFiles[i])
			if err != nil {
				t.Fail(err)

				return
			}

			updates = append(updates, notify.NewFileNotification(ctx, fInfo, websocket.FileUpdatedEvent, o)...)
		}

		ctx.Notify(ctx, updates...)
	}

	if err != nil {
		t.ReqNoErr(err)
	}

	pool.SignalAllQueued()

	// err = meta.FileService.ResizeDown(meta.File, nil, meta.Caster)
	// if err != nil {
	// 	log.Error().Stack().Err(err).Msg("")
	// }

	pool.Wait(true, t)

	errs := pool.Errors()
	if len(errs) != 0 {
		// Let any client subscribers know we failed
		result := task.Result{
			"failedCount": len(errs),
		}
		taskNotif := notify.NewTaskNotification(t, websocket.TaskFailedEvent, result)
		ctx.Notify(ctx, taskNotif)

		poolNotif := notify.NewPoolNotification(pool, websocket.TaskFailedEvent, result)
		ctx.Notify(ctx, poolNotif)

		t.Fail(wlerrors.WithStack(task.ErrChildTaskFailed))

		return
	}

	// Let any client subscribers know we are done
	result := getScanResult(t)
	notif := notify.NewPoolNotification(pool.GetRootPool(), websocket.FolderScanCompleteEvent, result)
	ctx.Notify(ctx, notif)

	t.Log().Debug().Func(func(e *zerolog.Event) { e.Msgf("Finished directory scan for %s", meta.File.GetPortablePath()) })

	t.Success()
}

// ScanFile scans an individual file and processes its metadata.
func ScanFile(tsk *task.Task) {
	reportSubscanStatus(tsk)

	meta := tsk.GetMeta().(job.ScanMeta)

	ctx, ok := context_service.FromContext(tsk.Ctx)
	if !ok {
		tsk.Fail(wlerrors.New("failed to get context"))

		return
	}

	tsk.SetResult(task.Result{
		"filepath": meta.File.GetPortablePath().String(),
	})

	err := ScanFileTsk(ctx, meta)
	if err != nil {
		tsk.Fail(err)
	}

	tsk.Success()
}

// ScanFileTsk is the internal implementation for scanning a file with the given context and metadata.
func ScanFileTsk(ctx context_service.AppContext, meta job.ScanMeta) error {
	if !media_model.ParseExtension(meta.File.GetPortablePath().Ext()).Displayable {
		return wlerrors.WithStack(media_model.ErrNotDisplayable)
	}

	cnf, err := featureflags.GetFlags(ctx)
	if err != nil {
		return wlerrors.WithStack(err)
	}

	existingMedia, err := media_model.GetMediaByContentID(ctx, meta.File.GetContentID())
	if err == nil && existingMedia.IsSufficentlyProcessed(cnf.EnableHDIR) {
		ctx.Log().Trace().Msgf("Media [%s] already sufficiently processed, skipping", existingMedia.ID())

		if !slices.Contains(existingMedia.FileIDs, meta.File.ID()) {
			err = existingMedia.AddFileToMedia(ctx, meta.File.ID())
			if err != nil {
				return err
			}
		}

		return nil
	}

	media := existingMedia
	mediaIsNew := media == nil
	isCached := false

	if mediaIsNew {
		media, err = media_service.NewMediaFromFile(ctx, meta.File)
		if err != nil {
			return err
		}
	} else {
		if !slices.Contains(existingMedia.FileIDs, meta.File.ID()) {
			err = existingMedia.AddFileToMedia(ctx, meta.File.ID())
			if err != nil {
				return err
			}
		}

		isCached, err = media_service.IsCached(ctx, media)
		if err != nil {
			return err
		}
	}

	if !isCached {
		_, err = media_service.HandleCacheCreation(ctx, media, meta.File)
		if err != nil {
			return err
		}
	}

	if len(media.HDIR) == 0 && cnf.EnableHDIR {
		_, err = media_service.GetHighDimensionImageEncoding(ctx, media)
		if err != nil {
			ctx.Log().Error().Err(err).Msgf("Failed to get HDIR encoding for %s", media.ID())
		}
	}

	ctx.Log().Trace().Func(func(e *zerolog.Event) {
		if !cnf.EnableHDIR {
			e.Msgf("HDIR generation is disabled, skipping for media %s", media.ID())
		}
	})

	err = media_model.SaveMedia(ctx, media)
	if err != nil {
		return err
	}

	mediaInfo := reshape.MediaToMediaInfo(media)

	o := notify.FileNotificationOptions{MediaInfo: mediaInfo}

	fInfo, err := reshape.WeblensFileToFileInfo(ctx, meta.File)
	if err != nil {
		return err
	}

	notif := notify.NewFileNotification(ctx, fInfo, websocket.FileUpdatedEvent, o)
	ctx.Notify(ctx, notif...)

	return nil
}

func queueScanFileIfNeeded(ctx context_service.AppContext, t *task.Task, mf *file_model.WeblensFileImpl, doHdir bool, alreadyFiles *[]*file_model.WeblensFileImpl, alreadyMedia *[]*media_model.Media, pool *task.Pool) error {
	if mf.IsDir() {
		t.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping file %s, not regular file", mf.GetPortablePath()) })

		return nil
	}

	if file_model.IsFileInTrash(mf) {
		t.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping file %s, file is in trash", mf.GetPortablePath()) })

		return nil
	}

	mt := media_model.ParseExtension(mf.GetPortablePath().Ext())
	if !mt.Displayable {
		return nil
	}

	if mf.GetContentID() == "" {
		t.Log().Error().Msgf("Skipping file %s, no content id", mf.GetPortablePath())

		return nil
	}

	m, err := media_model.GetMediaByContentID(ctx, mf.GetContentID())
	if err == nil && m.IsSufficentlyProcessed(doHdir) {
		if !slices.Contains(m.FileIDs, mf.ID()) {
			err = m.AddFileToMedia(ctx, mf.ID())
			if err != nil {
				return err
			}

			*alreadyFiles = append(*alreadyFiles, mf)
			*alreadyMedia = append(*alreadyMedia, m)
		}

		t.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping file %s, already imported", mf.GetPortablePath()) })

		return nil
	}

	subMeta := job.ScanMeta{
		File: mf,
	}

	t.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Dispatching scanFile job for [%s]", mf.GetPortablePath()) })

	newT, err := ctx.DispatchJob(job.ScanFileTask, subMeta, pool)
	if err != nil {
		return err
	}

	newT.SetCleanup(reportSubscanStatus)

	return nil
}

func reportSubscanStatus(tsk *task.Task) {
	event := websocket.FileScanStartedEvent

	if complete, exitStat := tsk.Status(); complete {
		switch exitStat {
		case task.TaskSuccess:
			event = websocket.FileScanCompleteEvent
		case task.TaskError:
			event = websocket.FileScanFailedEvent
		case task.TaskCanceled:
			event = websocket.FileScanCancelledEvent
		}
	}

	var notif websocket.WsResponseInfo
	if tsk.GetTaskPool().IsGlobal() || tsk.GetTaskPool().CreatedInTask() == nil {
		notif = notify.NewTaskNotification(tsk, event, getScanResult(tsk))
	} else {
		notif = notify.NewPoolNotification(tsk.GetTaskPool(), event, getScanResult(tsk))
	}

	ctx, ok := context_service.FromContext(tsk.Ctx)
	if !ok {
		tsk.Fail(wlerrors.New("failed to get context"))

		return
	}

	ctx.Notify(tsk.Ctx, notif)
}

func getScanResult(t *task.Task) task.Result {
	var tp *task.Pool

	if t.GetTaskPool() != nil {
		tp = t.GetTaskPool().GetRootPool()
	}

	var result = task.Result{}

	meta, ok := t.GetMeta().(job.ScanMeta)
	if ok {
		result = task.Result{
			"filename": meta.File.GetPortablePath().Filename(),
			"fileID":   meta.File.ID(),
		}

		createdIn := tp.CreatedInTask()
		if tp != nil && createdIn != nil {
			result["taskJobTarget"] = createdIn.GetMeta().(job.ScanMeta).File.GetPortablePath().Filename()
		} else if tp == nil {
			result["taskJobTarget"] = meta.File.GetPortablePath().Filename()
		}
	}

	if tp != nil {
		status := tp.Status()
		result["percentProgress"] = status.Progress
		result["tasksComplete"] = status.Complete
		result["tasksFailed"] = status.Failed
		result["tasksTotal"] = status.Total
		result["runtime"] = t.ExeTime()

		if tp.CreatedInTask() != nil {
			result["taskJobName"] = tp.CreatedInTask().JobName()
		}
	} else {
		result["taskJobName"] = t.JobName()
	}

	return result
}
