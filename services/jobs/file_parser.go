package jobs

import (
	"slices"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/job"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/models/task"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/services/context"
	file_service "github.com/ethanrous/weblens/services/file"
	media_service "github.com/ethanrous/weblens/services/media"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func ScanDirectory(tsk task_mod.Task) {
	t := tsk.(*task.Task)
	ctx := t.Ctx.(context.AppContext)
	meta := t.GetMeta().(job.ScanMeta)

	t.SetErrorCleanup(func(tsk task_mod.Task) {
		err := t.ReadError()
		notif := notify.NewTaskNotification(tsk.(*task.Task), websocket.TaskFailedEvent, task_mod.TaskResult{"error": err.Error()})
		t.Ctx.Notify(notif)
	})

	if file_service.IsFileInTrash(meta.File) {
		// Let any client subscribers know we are done
		notif := notify.NewTaskNotification(t, websocket.FolderScanCompleteEvent, task_mod.TaskResult{"executionTime": t.ExeTime()})
		t.Ctx.Notify(notif)
		t.Success("No media to scan")
		return
	}

	// TODO:
	// Claim task lock on this file before reading. This
	// prevents lost scans on child files if we were, say,
	// uploading into this directory as a scan comes through.
	// We will block until the upload finishes, then continue this scan
	// meta.File.AddTask(t)
	// defer func(meta.File *file_model.WeblensFileImpl, id task.TaskId) {
	// 	err := meta.File.RemoveTask(id)
	// 	if err != nil {
	// 		wlog.Error().Stack().Err(err).Msg("")
	// 	}
	// }(meta.File, t.Id())

	// Create a new task pool for the file scans this directory scan will spawn
	pool := t.GetTaskPool().GetWorkerPool().NewTaskPool(true, t)

	// err := meta.FileService.AddTask(meta.File, t)
	// t.ReqNoErr(err)
	//
	// defer func() {
	// 	err = meta.FileService.RemoveTask(meta.File, t)
	// 	if err != nil {
	// 		t.Log.Error().Stack().Err(err).Msg("")
	// 	}
	// }()

	ctx.ClientService.FolderSubToTask(ctx, meta.File.ID(), t)
	ctx.ClientService.FolderSubToTask(ctx, meta.File.GetParentId(), t)
	t.SetCleanup(func(tsk task_mod.Task) {
		// Make sure we finish sending any messages to the client
		// before we close unsubscribe from the task
		// meta.Caster.Flush()
		ctx.ClientService.UnsubTask(ctx, tsk.Id())
		// meta.Caster.Close()
	})

	t.Ctx.Log().Debug().Func(func(e *zerolog.Event) {
		e.Msgf("Beginning directory scan for %s (%s)", meta.File.GetPortablePath(), meta.File.ID())
	})

	var alreadyFiles []*file_model.WeblensFileImpl
	var alreadyMedia []*media_model.Media
	start := time.Now()
	err := meta.File.LeafMap(
		func(mf *file_model.WeblensFileImpl) error {
			if mf.IsDir() {
				log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping file %s, not regular file", mf.GetPortablePath()) })
				return nil
			}

			if file_service.IsFileInTrash(mf) {
				log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping file %s, file is in trash", mf.GetPortablePath()) })
				return nil
			}

			mt := media_model.ParseExtension(mf.GetPortablePath().Ext())
			if !mt.Displayable {
				return nil
			}

			media_model.GetMediaByContentId(ctx, mf.GetContentId())

			m, err := media_model.GetMediaByContentId(ctx, mf.GetContentId())
			if m != nil && m.IsImported() {
				if !slices.ContainsFunc(m.FileIDs, func(fId string) bool { return fId == mf.ID() }) {
					err = m.AddFileToMedia(ctx, mf.GetPortablePath().ToPortable())
					if err != nil {
						return err
					}

					alreadyFiles = append(alreadyFiles, mf)
					alreadyMedia = append(alreadyMedia, m)
				}
				log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping file %s, already imported", mf.GetPortablePath()) })
				return nil
			}

			subMeta := job.ScanMeta{
				File: mf,
			}
			log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Dispatching scanFile job for [%s]", mf.GetPortablePath()) })

			newT, err := t.Ctx.DispatchJob(job.ScanFileTask, subMeta, pool)
			if err != nil {
				return err
			}

			newT.SetCleanup(reportSubscanStatus)

			return nil
		},
	)

	log.Debug().Func(func(e *zerolog.Event) {
		e.Str("portable_file_path", meta.File.GetPortablePath().String()).Msgf("Directory scan found files in %s", time.Since(start))
	})

	// If the files are already in the media service, we need to update the clients
	// that may be waiting for these files to be processed, but since we won't be
	// adding those via the media service, we need to update the clients now
	if len(alreadyFiles) > 0 {
		updates := make([]websocket.WsResponseInfo, len(alreadyFiles))
		for i := range alreadyFiles {

			mediaInfo := reshape.MediaToMediaInfo(alreadyMedia[i])
			updates = append(updates, notify.NewFileNotification(ctx, alreadyFiles[i], websocket.FileUpdatedEvent, mediaInfo)...)
		}

		t.Ctx.Notify(updates...)
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
		t.Ctx.Notify(taskNotif)

		poolNotif := notify.NewPoolNotification(pool, websocket.TaskFailedEvent, result)
		t.Ctx.Notify(poolNotif)

		t.Fail(errors.WithStack(task_mod.ErrChildTaskFailed))

		return
	}

	// Let any client subscribers know we are done
	result := getScanResult(t)
	notif := notify.NewPoolNotification(pool.GetRootPool(), websocket.FolderScanCompleteEvent, result)
	t.Ctx.Notify(notif)

	t.Ctx.Log().Debug().Func(func(e *zerolog.Event) { e.Msgf("Finished directory scan for %s", meta.File.GetPortablePath()) })

	t.Success()
}

func ScanFile(tsk task_mod.Task) {
	t := tsk.(*task.Task)

	reportSubscanStatus(t)

	meta := t.GetMeta().(job.ScanMeta)
	err := ScanFile_(t.Ctx.(context.AppContext), meta)
	if err != nil {
		t.Ctx.Log().Error().Msgf("Failed to scan file %s: %s", meta.File.GetPortablePath(), err)
		t.Fail(err)
	}

	t.Success()
}

func ScanFile_(ctx context.AppContext, meta job.ScanMeta) error {
	ctx.WithLogger(ctx.Log().With().Str("file_id", meta.File.ID()).Str("portable_file_path", meta.File.GetPortablePath().String()).Logger())

	if !media_model.ParseExtension(meta.File.GetPortablePath().Ext()).Displayable {
		return media_model.ErrNotDisplayable
	}

	// contentId := meta.File.GetContentId()
	// if contentId == "" {
	// 	return errors.Errorf("trying to scan file with no content id: %s", meta.File.GetPortablePath())
	// }
	//
	// media := media_model.Media{ContentID: contentId}
	// if slices.ContainsFunc(
	// 	media.GetFiles(), func(fId string) bool {
	// 		return fId == meta.File.ID()
	// 	},
	// ) {
	// 	ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Media already exists for %s", meta.File.GetPortablePath()) })
	// 	return nil
	// }
	//
	// if meta.PartialMedia == nil {
	// 	meta.PartialMedia = &media_model.Media{}
	// }
	//
	// meta.PartialMedia.ContentID = meta.File.GetContentId()
	// meta.PartialMedia.FileIDs = []string{meta.File.ID()}
	//
	// username, err := user.GetFileOwnerName(ctx, meta.File)
	// if err != nil {
	// 	return err
	// }
	// meta.PartialMedia.Owner = username

	media, err := media_service.NewMediaFromFile(ctx, meta.File)
	if err != nil {
		return err
	}

	err = media_model.SaveMedia(ctx, media)
	if err != nil {
		return err
	}

	// existingMedia := meta.MediaService.Get(meta.PartialMedia.ID())
	// if existingMedia == nil || existingMedia.Height != meta.PartialMedia.Height || existingMedia.
	// 	Width != meta.PartialMedia.Width || len(existingMedia.FileIDs) != len(meta.PartialMedia.FileIDs) {
	// 	err = meta.MediaService.Add(meta.PartialMedia)
	// 	if err != nil && !errors.Is(err, media_model.ErrMediaAlreadyExists) {
	// 		return err
	// 	}
	// 	logger.Trace().Func(func(e *zerolog.Event) { e.Msgf("Added %s to media service", meta.File.Filename()) })
	// } else {
	// 	logger.Debug().Func(func(e *zerolog.Event) { e.Msgf("Media already exists for %s", meta.File.Filename()) })
	// }

	mediaInfo := reshape.MediaToMediaInfo(media)
	notif := notify.NewFileNotification(ctx, meta.File, websocket.FileUpdatedEvent, mediaInfo)
	ctx.Notify(notif...)

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
	tsk.Ctx.Notify(notif)
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
