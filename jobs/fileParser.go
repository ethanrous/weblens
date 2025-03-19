package jobs

import (
	"errors"
	"path/filepath"
	"slices"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/caster"
	"github.com/ethanrous/weblens/task"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func ScanDirectory(t *task.Task) {
	meta := t.GetMeta().(models.ScanMeta)
	meta.Caster = caster.NewSimpleCaster(meta.TaskSubber.(models.ClientManager), t.Log)

	t.SetErrorCleanup(func(tsk *task.Task) {
		err := t.ReadError()
		meta.Caster.PushTaskUpdate(tsk, models.TaskFailedEvent, task.TaskResult{"error": err.Error()})
	})

	if meta.FileService.IsFileInTrash(meta.File) {
		// Let any client subscribers know we are done
		meta.Caster.PushTaskUpdate(
			t, models.FolderScanCompleteEvent, task.TaskResult{"executionTime": t.ExeTime()},
		)
		t.Success("No media to scan")
		return
	}

	// TODO:
	// Claim task lock on this file before reading. This
	// prevents lost scans on child files if we were, say,
	// uploading into this directory as a scan comes through.
	// We will block until the upload finishes, then continue this scan
	// meta.File.AddTask(t)
	// defer func(meta.File *fileTree.WeblensFileImpl, id task.TaskId) {
	// 	err := meta.File.RemoveTask(id)
	// 	if err != nil {
	// 		wlog.Error().Stack().Err(err).Msg("")
	// 	}
	// }(meta.File, t.TaskId())

	// Create a new task pool for the file scans this directory scan will spawn
	pool := t.GetTaskPool().GetWorkerPool().NewTaskPool(true, t)

	err := meta.FileService.AddTask(meta.File, t)
	t.ReqNoErr(err)

	defer func() {
		err = meta.FileService.RemoveTask(meta.File, t)
		if err != nil {
			t.Log.Error().Stack().Err(err).Msg("")
		}
	}()

	meta.TaskSubber.FolderSubToTask(meta.File.ID(), t.TaskId())
	meta.TaskSubber.FolderSubToTask(meta.File.GetParentId(), t.TaskId())
	t.SetCleanup(func(tsk *task.Task) {
		// Make sure we finish sending any messages to the client
		// before we close unsubscribe from the task
		meta.Caster.Flush()
		meta.TaskSubber.UnsubTask(tsk.TaskId())
		meta.Caster.Close()
	})

	t.Log.Debug().Func(func(e *zerolog.Event) {
		e.Msgf("Beginning directory scan for %s (%s)", meta.File.GetPortablePath(), meta.File.ID())
	})

	var alreadyFiles []*fileTree.WeblensFileImpl
	var alreadyMedia []*models.Media
	start := time.Now()
	err = meta.File.LeafMap(
		func(mf *fileTree.WeblensFileImpl) error {
			if mf.IsDir() {
				log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping file %s, not regular file", mf.GetPortablePath()) })
				return nil
			}

			if meta.FileService.IsFileInTrash(mf) {
				log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping file %s, file is in trash", mf.GetPortablePath()) })
				return nil
			}

			if !meta.MediaService.IsFileDisplayable(mf) {
				log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping file %s, not displayable", mf.GetPortablePath()) })
				return nil
			}

			m := meta.MediaService.Get(mf.GetContentId())
			if m != nil && m.IsImported() && meta.MediaService.IsCached(m) {
				if !slices.ContainsFunc(m.FileIDs, func(fId fileTree.FileId) bool { return fId == mf.ID() }) {
					err := meta.MediaService.AddFileToMedia(m, mf)
					if err != nil {
						return err
					}
					alreadyFiles = append(alreadyFiles, mf)
					alreadyMedia = append(alreadyMedia, m)
				}
				log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Skipping file %s, already imported", mf.GetPortablePath()) })
				return nil
			}

			subMeta := models.ScanMeta{
				File:         mf,
				FileService:  meta.FileService,
				MediaService: meta.MediaService,
				Caster:       meta.Caster,
			}
			log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Dispatching scanFile job for [%s]", mf.GetPortablePath()) })
			newT, err := meta.TaskService.DispatchJob(models.ScanFileTask, subMeta, pool)
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
	// adding those to the media service, we need to update the clients now
	meta.Caster.PushFilesUpdate(alreadyFiles, alreadyMedia)

	if err != nil {
		t.ReqNoErr(err)
	}

	pool.SignalAllQueued()

	err = meta.FileService.ResizeDown(meta.File, nil, meta.Caster)
	if err != nil {
		log.Error().Stack().Err(err).Msg("")
	}

	pool.Wait(true, t)

	errs := pool.Errors()
	if len(errs) != 0 {
		// Let any client subscribers know we failed
		meta.Caster.PushTaskUpdate(
			t, models.TaskFailedEvent, task.TaskResult{
				"failedCount": len(errs),
			},
		)
		meta.Caster.PushPoolUpdate(
			pool, models.TaskFailedEvent, task.TaskResult{
				"failedCount": len(errs),
			},
		)
		t.Fail(werror.WithStack(werror.ErrChildTaskFailed))
	}

	// Let any client subscribers know we are done
	result := getScanResult(t)
	meta.Caster.PushPoolUpdate(pool.GetRootPool(), models.FolderScanCompleteEvent, result)

	t.Log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Finished directory scan for %s", meta.File.GetPortablePath()) })

	t.Success()
}

func ScanFile(t *task.Task) {
	reportSubscanStatus(t)

	meta := t.GetMeta().(models.ScanMeta)
	err := ScanFile_(meta, t.ExitIfSignaled, t.Log)
	if err != nil {
		t.Log.Error().Msgf("Failed to scan file %s: %s", meta.File.GetPortablePath(), err)
		t.Fail(err)
	}

	t.Success()
}

func ScanFile_(meta models.ScanMeta, exitCheck func(), logger *zerolog.Logger) error {
	logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("portable_file_path", meta.File.GetPortablePath().String()).Str("file_id", meta.File.ID())
	})

	sw := internal.NewStopwatch("Scan "+meta.File.Filename(), logger)
	ext := filepath.Ext(meta.File.Filename())
	if !meta.MediaService.GetMediaTypes().ParseExtension(ext).Displayable {
		logger.Error().Msgf("Trying to process file with [%s] ext", ext)
		return werror.ErrNonDisplayable
	}
	sw.Lap("Check ext")

	contentId := meta.File.GetContentId()
	if contentId == "" {
		return werror.Errorf("trying to scan file with no content id: %s", meta.File.GetPortablePath())
	}
	sw.Lap("Check contentId")

	media := models.NewMedia(contentId)
	if slices.ContainsFunc(
		media.GetFiles(), func(fId fileTree.FileId) bool {
			return fId == meta.File.ID()
		},
	) {
		logger.Trace().Func(func(e *zerolog.Event) { e.Msgf("Media already exists for %s", meta.File.Filename()) })
		return nil
	}
	sw.Lap("New media")

	if meta.PartialMedia == nil {
		meta.PartialMedia = &models.Media{}
	}

	meta.PartialMedia.ContentID = meta.File.GetContentId()
	meta.PartialMedia.FileIDs = []fileTree.FileId{meta.File.ID()}

	owner, err := meta.FileService.GetFileOwner(meta.File)
	if err != nil {
		return err
	}
	meta.PartialMedia.Owner = owner.GetUsername()

	exitCheck()

	sw.Lap("Checked metadata")
	err = meta.MediaService.LoadMediaFromFile(meta.PartialMedia, meta.File)
	if err != nil {
		return err
	}
	sw.Lap("Loaded media from file")

	exitCheck()

	existingMedia := meta.MediaService.Get(meta.PartialMedia.ID())
	if existingMedia == nil || existingMedia.Height != meta.PartialMedia.Height || existingMedia.
		Width != meta.PartialMedia.Width || len(existingMedia.FileIDs) != len(meta.PartialMedia.FileIDs) {
		err = meta.MediaService.Add(meta.PartialMedia)
		if err != nil && !errors.Is(err, werror.ErrMediaAlreadyExists) {
			return err
		}
		logger.Trace().Func(func(e *zerolog.Event) { e.Msgf("Added %s to media service", meta.File.Filename()) })
		sw.Lap("Added media to service")
	} else {
		logger.Debug().Func(func(e *zerolog.Event) { e.Msgf("Media already exists for %s", meta.File.Filename()) })
	}

	meta.Caster.PushFileUpdate(meta.File, meta.PartialMedia)
	sw.Lap("Pushed updated file")
	sw.Stop()
	// sw.PrintResults(false)
	logger.Trace().Func(func(e *zerolog.Event) { e.Msgf("Finished processing %s", meta.File.Filename()) })

	return nil
}

func reportSubscanStatus(t *task.Task) {
	meta := t.GetMeta().(models.ScanMeta)
	event := models.FileScanStartedEvent
	if complete, _ := t.Status(); complete {
		event = models.FileScanCompleteEvent
	}

	if t.GetTaskPool().IsGlobal() {
		meta.Caster.PushTaskUpdate(t, event, getScanResult(t))
	} else {
		meta.Caster.PushPoolUpdate(t.GetTaskPool().GetRootPool(), event, getScanResult(t))
	}
}

func getScanResult(t *task.Task) task.TaskResult {
	var tp *task.TaskPool

	if t.GetTaskPool() != nil {
		tp = t.GetTaskPool().GetRootPool()
	}

	var result = task.TaskResult{}
	meta, ok := t.GetMeta().(models.ScanMeta)
	if ok {
		result = task.TaskResult{
			"filename": meta.File.Filename(),
			"fileId":   meta.File.ID(),
		}
		if tp != nil && tp.CreatedInTask() != nil {
			result["taskJobTarget"] = tp.CreatedInTask().GetMeta().(models.ScanMeta).File.Filename()
		} else if tp == nil {
			result["taskJobTarget"] = meta.File.Filename()
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
