package jobs

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"slices"
	"strings"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/models/job"
	job_model "github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/rs/zerolog"
)

func removeTopLevels(t *task.Task, topLevels []*file_model.WeblensFileImpl) error {
	for _, tl := range topLevels {
		t.Log().Debug().Msgf("Top level to remove: %s", tl.GetPortablePath())
	}

	appCtx, ok := context_service.FromContext(t.Ctx)
	if !ok {
		return wlerrors.New("failed to get context")
	}

	ctx := context.WithoutCancel(t.Ctx)

	err := appCtx.FileService.DeleteFiles(ctx, topLevels...)
	if err != nil {
		return err
	}

	return nil
}

// HandleFileUploads is the job for reading file chunks coming in from client requests, and writing them out
// to their corresponding files. Intention behind this implementation is to rid the
// client of interacting with any blocking calls to make the upload process as fast as
// possible, hopefully as fast as the slower of the 2 network speeds. This task handles
// everything *after* the client has had its data read into memory, this is the "bottom half"
// of the upload
func HandleFileUploads(tsk task_mod.Task) {
	t := tsk.(*task.Task)

	appCtx, ok := context_service.FromContext(t.Ctx)
	if !ok {
		t.Fail(wlerrors.New("failed to get context"))

		return
	}

	meta := t.GetMeta().(job.UploadFilesMeta)

	rootFile, err := appCtx.FileService.GetFileByID(appCtx, meta.RootFolderID)
	if err != nil {
		t.Fail(err)

		return
	}

	// This map will only be accessed by this task and by this 1 thread,
	// so we do not need any synchronization here
	fileMap := map[string]*job_model.FileUploadProgress{}
	usingFiles := []string{}
	topLevels := []*file_model.WeblensFileImpl{}
	timeout := false

	// Cleanup routine. This must be run even if the upload fails
	t.SetCleanup(func(tsk task_mod.Task) {
		t := tsk.(*task.Task)

		for _, f := range fileMap {
			t.Log().Debug().Func(func(e *zerolog.Event) {
				e.Msgf("Cleaning up file [%+v]", *f)
			})
		}
		// e.Msgf("Upload fileMap has %d remaining - and chunk stream has %d remaining", len(fileMap), len(meta.ChunkStream))

		select {
		case <-t.Ctx.Done():
			err = removeTopLevels(t, topLevels)
			if err != nil {
				t.Log().Error().Stack().Err(err).Msg("Failed to remove top level files")
			}

			return
		default:
		}

		rootFile.Size()
		notifs := notify.NewFileNotification(appCtx, rootFile, websocket.FileUpdatedEvent)

		doingRootScan := false
		// Do not report that this task pool was created by this task, we want to detach
		// and allow these scans to take place independently
		newTp := t.GetTaskPool().GetWorkerPool().NewTaskPool(false, nil)

		for _, tl := range topLevels {
			notif := notify.NewFileNotification(appCtx, tl, websocket.FileUpdatedEvent)
			notifs = append(notifs, notif...)

			if tl.IsDir() {
				if !timeout {
					scanMeta := job_model.ScanMeta{
						File: tl,
					}

					_, err = appCtx.DispatchJob(job_model.ScanDirectoryTask, scanMeta, newTp)
					if err != nil {
						t.Log().Error().Stack().Err(err).Msg("")

						continue
					}
				}
			} else if !doingRootScan && !timeout {
				scanMeta := job_model.ScanMeta{
					File: rootFile,
				}

				_, err = appCtx.DispatchJob(job_model.ScanDirectoryTask, scanMeta, newTp)
				if err != nil {
					t.Log().Error().Stack().Err(err).Msg("")

					continue
				}

				doingRootScan = true
			}
		}

		appCtx.Notify(appCtx, notifs...)
		newTp.SignalAllQueued()
	})

	timeoutTicker := time.NewTicker(time.Minute)

	ctx := history.WithFileEvent(t.Ctx)

	appCtx, ok = context_service.FromContext(ctx)
	if !ok {
		t.Fail(wlerrors.New("failed to add file event to context in upload task"))

		return
	}

WriterLoop:
	for {
		timeoutTicker.Reset(time.Minute)

		select {
		case <-t.Ctx.Done(): // Listen for cancellation
			t.Log().Debug().Func(func(e *zerolog.Event) { e.Msgf("Context done, exiting upload task") })
			t.Fail(task.ErrTaskCancelled)

			return
		case <-timeoutTicker.C:
			t.Log().Debug().Func(func(e *zerolog.Event) { e.Msgf("Timeout, exiting upload task") })
			t.Fail(task.ErrTaskTimeout)

			timeout = true

			return
		case chunk := <-meta.ChunkStream:
			bottom, top, total, err := parseRangeHeader(chunk.ContentRange)
			t.ReqNoErr(err)

			if chunk.NewFile != nil {
				tmpFile := chunk.NewFile
				for tmpFile.GetParent() != rootFile {
					tmpFile = tmpFile.GetParent()
				}

				if !slices.ContainsFunc(topLevels, func(tl *file_model.WeblensFileImpl) bool { return tl.ID() == tmpFile.ID() }) {
					topLevels = append(topLevels, tmpFile)
				}

				fileMap[chunk.NewFile.ID()] = &job_model.FileUploadProgress{
					File: chunk.NewFile, BytesWritten: 0, FileSizeTotal: total, Hash: sha256.New(),
				}

				slices_mod.InsertFunc(
					usingFiles, chunk.NewFile.ID(),
					func(a, b string) int { return strings.Compare(string(a), string(b)) },
				)

				t.Log().Trace().Func(func(e *zerolog.Event) {
					e.Msgf("New upload [%s] of size [%d bytes]", chunk.NewFile.GetPortablePath(), total)
				})

				continue WriterLoop
			}

			// We use `0-0/-1` as a fake "range header" to indicate that the upload for
			// the specific file has had an error or been canceled, and should be removed.
			if total == -1 {
				delete(fileMap, chunk.FileID)

				continue
			}

			chnk := fileMap[chunk.FileID]

			if chnk.FileSizeTotal != total {
				t.Fail(wlerrors.Errorf("upload size mismatch for file [%s / %s] (%d != %d)", chnk.File.GetPortablePath(), chnk.File.ID(), chnk.FileSizeTotal, total))
			}

			// Add the new bytes to the counter for the file-size of this file.
			// If we upload content in range e.g. 0-1 bytes, that includes 2 bytes,
			// but top - bottom (1 - 0) is 1, so we add 1 to match
			chnk.BytesWritten += (top - bottom) + 1

			// Write the bytes to the real file
			err = chnk.File.WriteAt(chunk.Chunk, bottom)
			if err != nil {
				t.ReqNoErr(err)
			}

			// Add the bytes for this chunk to the Hash
			_, err = chnk.Hash.Write(chunk.Chunk)
			if err != nil {
				t.ReqNoErr(err)
			}

			// When file is finished writing
			if chnk.BytesWritten >= chnk.FileSizeTotal {
				t.Log().Debug().Func(func(e *zerolog.Event) {
					e.Msgf("Finished writing file [%s] with %d bytes", chnk.File.GetPortablePath(), chnk.BytesWritten)
				})

				// Hash file content to get content ID.
				chnk.File.SetContentID(base64.URLEncoding.EncodeToString(chnk.Hash.Sum(nil))[:20])

				if chnk.File.GetContentID() == "" {
					t.Fail(wlerrors.Errorf("failed to generate contentID for file upload [%s]", chnk.File.GetPortablePath()))
				}

				newAction := history.NewCreateAction(appCtx, chnk.File)

				err = history.SaveAction(appCtx, &newAction)
				if err != nil {
					t.Fail(err)

					return
				}

				notif := notify.NewFileNotification(appCtx, chnk.File, websocket.FileCreatedEvent)
				appCtx.Notify(appCtx, notif...)

				// Remove the file from our local map
				i, e := slices.BinarySearchFunc(
					usingFiles, chunk.FileID,
					func(a, b string) int { return strings.Compare(a, b) },
				)
				if e {
					usingFiles = slices.Delete(usingFiles, i, i+1)
				}

				delete(fileMap, chunk.FileID)
			}

			t.Log().Trace().Func(func(e *zerolog.Event) {
				if chnk.BytesWritten < chnk.FileSizeTotal {
					e.Msgf("%s has not finished uploading yet %d of %d", chnk.File.GetPortablePath(), chnk.BytesWritten, chnk.FileSizeTotal)
				}
			})

			// When we have no files being worked on, and there are no more
			// chunks to write, we are finished.
			if len(fileMap) == 0 && len(meta.ChunkStream) == 0 {
				break WriterLoop
			}

			continue WriterLoop
		}
	}

	t.Log().Debug().Func(func(e *zerolog.Event) { e.Msgf("Finished writing upload files for %s", rootFile.GetPortablePath()) })
	t.Success()
}
