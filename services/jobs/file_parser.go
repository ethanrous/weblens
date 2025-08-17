package jobs

import (
	"slices"
	"time"

	"github.com/ethanrous/weblens/models/config"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/job"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/errors"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/ethanrous/weblens/modules/websocket"
	context_service "github.com/ethanrous/weblens/services/context"
	media_service "github.com/ethanrous/weblens/services/media"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func ScanDirectory(tsk task_mod.Task) {
	t := tsk.(*task.Task)
	ctx, ok := context_service.FromContext(t.Ctx)
	if !ok {
		t.Fail(errors.New("failed to get context"))

		return
	}
	meta := t.GetMeta().(job.ScanMeta)

	t.SetErrorCleanup(func(tsk task_mod.Task) {
		err := t.ReadError()
		notif := notify.NewTaskNotification(tsk.(*task.Task), websocket.TaskFailedEvent, task_mod.TaskResult{"error": err.Error()})
		ctx.Notify(ctx, notif)
	})

	if file_model.IsFileInTrash(meta.File) {
		// Let any client subscribers know we are done
		notif := notify.NewTaskNotification(t, websocket.FolderScanCompleteEvent, task_mod.TaskResult{"executionTime": t.ExeTime()})
		ctx.Notify(ctx, notif)
		t.Success("No media to scan")

		return
	}

	// Create a new task pool for the file scans this directory scan will spawn
	pool := t.GetTaskPool().GetWorkerPool().NewTaskPool(true, t)

	ctx.ClientService.FolderSubToTask(ctx, meta.File.ID(), t)
	ctx.ClientService.FolderSubToTask(ctx, meta.File.GetParent().ID(), t)
	t.SetCleanup(func(tsk task_mod.Task) {
		// Make sure we finish sending any messages to the client
		// before we close unsubscribe from the task
		// meta.Caster.Flush()
		ctx.ClientService.Flush(ctx)
		ctx.ClientService.UnsubTask(ctx, tsk.Id())
	})

	t.Log().Debug().Func(func(e *zerolog.Event) {
		e.Msgf("Beginning directory scan for %s (%s)", meta.File.GetPortablePath(), meta.File.ID())
	})

	var alreadyFiles []*file_model.WeblensFileImpl

	var alreadyMedia []*media_model.Media

	start := time.Now()

	cnf, err := config.GetConfig(ctx)
	if err != nil {
		t.Fail(errors.WithStack(err))

		return
	}

	err = meta.File.LeafMap(
		func(mf *file_model.WeblensFileImpl) error {
			return queueScanFileIfNeeded(ctx, t, mf, cnf.EnableHDIR, &alreadyFiles, &alreadyMedia, pool)
		},
	)

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
			updates = append(updates, notify.NewFileNotification(ctx, alreadyFiles[i], websocket.FileUpdatedEvent, o)...)
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
		result := task_mod.TaskResult{
			"failedCount": len(errs),
		}
		taskNotif := notify.NewTaskNotification(t, websocket.TaskFailedEvent, result)
		ctx.Notify(ctx, taskNotif)

		poolNotif := notify.NewPoolNotification(pool, websocket.TaskFailedEvent, result)
		ctx.Notify(ctx, poolNotif)

		t.Fail(errors.WithStack(task_mod.ErrChildTaskFailed))

		return
	}

	// Let any client subscribers know we are done
	result := getScanResult(t)
	notif := notify.NewPoolNotification(pool.GetRootPool(), websocket.FolderScanCompleteEvent, result)
	ctx.Notify(ctx, notif)

	t.Log().Debug().Func(func(e *zerolog.Event) { e.Msgf("Finished directory scan for %s", meta.File.GetPortablePath()) })

	t.Success()
}

func ScanFile(tsk task_mod.Task) {
	t := tsk.(*task.Task)

	reportSubscanStatus(t)

	meta := t.GetMeta().(job.ScanMeta)

	ctx, ok := context_service.FromContext(t.Ctx)
	if !ok {
		t.Fail(errors.New("failed to get context"))

		return
	}

	err := ScanFile_(ctx, meta)
	if err != nil {
		t.Fail(err)
	}

	t.Success()
}

func ScanFile_(ctx context_service.AppContext, meta job.ScanMeta) error {
	if !media_model.ParseExtension(meta.File.GetPortablePath().Ext()).Displayable {
		return errors.WithStack(media_model.ErrNotDisplayable)
	}

	cnf, err := config.GetConfig(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	existingMedia, err := media_model.GetMediaByContentId(ctx, meta.File.GetContentId())
	if err == nil && existingMedia.IsSufficentlyProcessed(cnf.EnableHDIR) {
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
		ctx.Log().Error().Err(err).Msgf("Failed to get HDIR encoding for %s", media.ID())
	}

	err = media_model.SaveMedia(ctx, media)
	if err != nil {
		return err
	}

	mediaInfo := reshape.MediaToMediaInfo(media)

	o := notify.FileNotificationOptions{MediaInfo: mediaInfo}
	notif := notify.NewFileNotification(ctx, meta.File, websocket.FileUpdatedEvent, o)
	ctx.Notify(ctx, notif...)

	return nil
}

func queueScanFileIfNeeded(ctx context_service.AppContext, t *task.Task, mf *file_model.WeblensFileImpl, doHdir bool, alreadyFiles *[]*file_model.WeblensFileImpl, alreadyMedia *[]*media_model.Media, pool task_mod.Pool) error {
	start := time.Now()

	if mf.IsDir() {
		t.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping file %s, not regular file", mf.GetPortablePath()) })

		t.Log().Debug().Msgf("Not scanned directory %s", time.Since(start))
		return nil
	}

	if file_model.IsFileInTrash(mf) {
		t.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping file %s, file is in trash", mf.GetPortablePath()) })

		t.Log().Debug().Msgf("Not scanned trash file %s", time.Since(start))
		return nil
	}

	mt := media_model.ParseExtension(mf.GetPortablePath().Ext())
	if !mt.Displayable {
		t.Log().Debug().Msgf("Not scanned non-displayable file %s", time.Since(start))

		return nil
	}

	if mf.GetContentId() == "" {
		t.Log().Error().Msgf("Skipping file %s, no content id", mf.GetPortablePath())

		t.Log().Debug().Msgf("Not scanned no content id %s", time.Since(start))
		return nil
	}

	m, err := media_model.GetMediaByContentId(ctx, mf.GetContentId())
	t.Log().Debug().Msgf("fetched media %s", time.Since(start))
	if err == nil && m.IsSufficentlyProcessed(doHdir) {
		if !slices.Contains(m.FileIDs, mf.ID()) {
			t.Log().Debug().Msgf("needed file id check %v %s %s", m.FileIDs, mf.ID(), time.Since(start))
			err = m.AddFileToMedia(ctx, mf.ID())
			t.Log().Debug().Msgf("needed file id %s", time.Since(start))
			if err != nil {
				return err
			}

			*alreadyFiles = append(*alreadyFiles, mf)
			*alreadyMedia = append(*alreadyMedia, m)
		}
		t.Log().Debug().Msgf("needed file id check after %s", time.Since(start))

		t.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping file %s, already imported", mf.GetPortablePath()) })

		t.Log().Debug().Msgf("Not scanned already imported %s", time.Since(start))
		return nil
	}

	t.Log().Debug().Msgf("not sufficently processed: %v+ %s", m, time.Since(start))

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

func reportSubscanStatus(t task_mod.Task) {
	event := websocket.FileScanStartedEvent
	if complete, _ := t.Status(); complete {
		event = websocket.FileScanCompleteEvent
	}

	tsk := t.(*task.Task)

	var notif websocket.WsResponseInfo
	if t.GetTaskPool().IsGlobal() {
		notif = notify.NewTaskNotification(tsk, event, getScanResult(tsk))
	} else {
		notif = notify.NewPoolNotification(tsk.GetTaskPool(), event, getScanResult(tsk))
	}

	ctx, ok := context_service.FromContext(tsk.Ctx)
	if !ok {
		tsk.Fail(errors.New("failed to get context"))

		return
	}

	ctx.Notify(ctx, notif)
}

func getScanResult(t *task.Task) task_mod.TaskResult {
	var tp *task.TaskPool

	if t.GetTaskPool() != nil {
		tp = t.GetTaskPool().GetRootPool().(*task.TaskPool)
	}

	var result = task_mod.TaskResult{}

	meta, ok := t.GetMeta().(job.ScanMeta)
	if ok {
		result = task_mod.TaskResult{
			"filename": meta.File.GetPortablePath().Filename(),
			"fileId":   meta.File.ID(),
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
