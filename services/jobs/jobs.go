package jobs

import (
	"crypto/sha256"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/pkg/errors"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/client"
	file_model "github.com/ethanrous/weblens/models/file"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	"github.com/ethanrous/weblens/service"
	"github.com/ethanrous/weblens/task"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func parseRangeHeader(contentRange string) (min, max, total int64, err error) {
	rangeAndSize := strings.Split(contentRange, "/")
	rangeParts := strings.Split(rangeAndSize[0], "-")

	min, err = strconv.ParseInt(rangeParts[0], 10, 64)
	if err != nil {
		return
	}

	max, err = strconv.ParseInt(rangeParts[1], 10, 64)
	if err != nil {
		return
	}

	total, err = strconv.ParseInt(rangeAndSize[1], 10, 64)
	if err != nil {
		return
	}
	return
}

// HandleFileUploads is the job for reading file chunks coming in from client requests, and writing them out
// to their corresponding files. Intention behind this implementation is to rid the
// client of interacting with any blocking calls to make the upload process as fast as
// possible, hopefully as fast as the slower of the 2 network speeds. This task handles
// everything *after* the client has had its data read into memory, this is the "bottom half"
// of the upload
func HandleFileUploads(t *task.Task) {
	meta := t.GetMeta().(models.UploadFilesMeta)

	t.ExitIfSignaled()

	rootFile, err := meta.FileService.GetFileSafe(meta.RootFolderId, meta.User, meta.Share)
	if err != nil {
		t.ReqNoErr(err)
	}

	// This map will only be accessed by this task and by this 1 thread,
	// so we do not need any synchronization here
	fileMap := map[file_model.FileId]*models.FileUploadProgress{}

	// meta.Caster.DisableAutoFlush()
	var usingFiles []file_model.FileId
	var topLevels []*file_model.WeblensFileImpl

	fileEvent := meta.UploadEvent

	timeout := false

	t.SetErrorCleanup(func(*task.Task) {
		if fileEvent.Logged.Load() {
			return
		}
		t.Log.Trace().Msg("Doing error cleanup journal log for upload")
		meta.FileService.GetJournalByTree("USERS").LogEvent(fileEvent)
	})

	// Cleanup routine. This must be run even if the upload fails
	t.SetCleanup(func(t *task.Task) {

		t.Log.Debug().Func(func(e *zerolog.Event) {
			e.Msgf("Upload fileMap has %d remaining and chunk stream has %d remaining", len(fileMap), len(meta.ChunkStream))
		})

		doingRootScan := false

		// Do not report that this task pool was created by this task, we want to detach
		// and allow these scans to take place independently
		newTp := t.GetTaskPool().GetWorkerPool().NewTaskPool(false, nil)
		for _, tl := range topLevels {
			if tl.IsDir() {
				err = meta.FileService.ResizeDown(tl, fileEvent, meta.Caster)
				if err != nil {
					t.ReqNoErr(err)
				}

				if !timeout {
					scanMeta := models.ScanMeta{
						File:         tl,
						FileService:  meta.FileService,
						TaskService:  meta.TaskService,
						MediaService: meta.MediaService,
						TaskSubber:   meta.TaskSubber,
					}
					_, err = t.GetTaskPool().GetWorkerPool().DispatchJob(models.ScanDirectoryTask, scanMeta, newTp)
					if err != nil {
						t.Log.Error().Stack().Err(err).Msg("")
						continue
					}
				}
			} else if !doingRootScan && !timeout {
				scanMeta := models.ScanMeta{
					File:         rootFile,
					FileService:  meta.FileService,
					TaskService:  meta.TaskService,
					MediaService: meta.MediaService,
					TaskSubber:   meta.TaskSubber,
				}
				_, err = t.GetTaskPool().GetWorkerPool().DispatchJob(models.ScanDirectoryTask, scanMeta, newTp)
				if err != nil {
					t.Log.Error().Stack().Err(err).Msg("")
					continue
				}
				doingRootScan = true
			}
			media := meta.MediaService.Get(tl.GetContentId())
			if tl.IsDir() {
				meta.Caster.PushFileUpdate(tl, media)
			}
		}
		newTp.SignalAllQueued()

		err = meta.FileService.ResizeUp(rootFile, fileEvent, meta.Caster)
		if err != nil {
			t.Log.Error().Stack().Err(err).Msg("")
		}

		meta.FileService.GetJournalByTree("USERS").LogEvent(fileEvent)
		fileEvent.Wait()

		if newTp.Status().Total != 0 {
			newTp.AddCleanup(
				func(_ task.Pool) {
					meta.Caster.Close()
				},
			)
		} else {
			meta.Caster.Close()
		}
	})

WriterLoop:
	for {
		t.SetTimeout(time.Now().Add(time.Second * 60))
		select {
		case signal := <-t.GetSignalChan(): // Listen for cancellation
			if signal == 1 {
				timeout = true
				break WriterLoop
			}
		case chunk := <-meta.ChunkStream:
			t.ClearTimeout()

			bottom, top, total, err := parseRangeHeader(chunk.ContentRange)
			t.ReqNoErr(err)

			if chunk.NewFile != nil {

				tmpFile := chunk.NewFile
				for tmpFile.GetParent() != rootFile {
					tmpFile = tmpFile.GetParent()
				}

				if tmpFile.GetParent() == rootFile {
					if _, ok := slices.BinarySearchFunc(
						topLevels,
						string(tmpFile.ID()),
						// TODO: Make sure that this is comparing in the correct order
						func(f *file_model.WeblensFileImpl, t string) int { return strings.Compare(t, f.ID()) },
					); ok {
						topLevels = append(topLevels, tmpFile)
					}
				}

				fileMap[chunk.NewFile.ID()] = &models.FileUploadProgress{
					File: chunk.NewFile, BytesWritten: 0, FileSizeTotal: total, Hash: sha256.New(),
				}

				slices_mod.InsertFunc(
					usingFiles, chunk.NewFile.ID(),
					func(a, b file_model.FileId) int { return strings.Compare(string(a), string(b)) },
				)

				t.Log.Trace().Func(func(e *zerolog.Event) {
					e.Msgf("New upload [%s] of size [%d bytes]", chunk.NewFile.GetPortablePath(), total)
				})

				continue WriterLoop
			}

			// We use `0-0/-1` as a fake "range header" to indicate that the upload for
			// the specific file has had an error or been canceled, and should be removed.
			if total == -1 {
				delete(fileMap, chunk.FileId)
				continue
			}

			chnk := fileMap[chunk.FileId]

			if chnk.FileSizeTotal != total {
				t.Fail(werror.Errorf("upload size mismatch for file [%s / %s] (%d != %d)", chnk.File.GetPortablePath(), chnk.File.ID(), chnk.FileSizeTotal, total))
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

				// Hash file content to get content ID. Must do this before attaching the file,
				// or the journal worker will beat us to it, which could break if importing
				// the file media shortly after uploading here.
				chnk.File.SetContentId(service.ContentIdFromHash(chnk.Hash))
				if chnk.File.GetContentId() == "" {
					t.Fail(werror.Errorf("failed to generate contentId for file upload [%s]", chnk.File.AbsPath()))
				}

				if !chnk.File.IsDir() {
					meta.Caster.PushFileCreate(chnk.File)
				}

				newAction := fileEvent.NewCreateAction(chnk.File)
				if newAction == nil {
					t.Fail(werror.Errorf("failed to create new file action on upload for [%s]", chnk.File.AbsPath()))
				}

				// Remove the file from our local map
				i, e := slices.BinarySearchFunc(
					usingFiles, chunk.FileId,
					func(a, b fileTree.FileId) int { return strings.Compare(string(a), string(b)) },
				)
				if e {
					usingFiles = slices.Delete(usingFiles, i, i+1)
				}
				delete(fileMap, chunk.FileId)
			}
			t.Log.Trace().Func(func(e *zerolog.Event) {
				if chnk.BytesWritten < chnk.FileSizeTotal {
					e.Msgf("%s has not finished uploading yet %d of %d", chnk.File.AbsPath(), chnk.BytesWritten, chnk.FileSizeTotal)
				}
			})

			// When we have no files being worked on, and there are no more
			// chunks to write, we are finished.
			if len(fileMap) == 0 && len(meta.ChunkStream) == 0 {
				break WriterLoop
			}
			t.ExitIfSignaled()
			continue WriterLoop
		}
	}

	t.Log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Finished writing upload files for %s", rootFile.GetPortablePath()) })
	t.Success()
}

type extSize struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

func GatherFilesystemStats(t *task.Task) {
	meta := t.GetMeta().(models.FsStatMeta)

	filetypeSizeMap := map[string]int64{}
	folderCount := 0

	// media := dataStore.GetMediaDir()
	// external := dataStore.GetExternalDir()
	// dataStore.ResizeDown(media)

	sizeFunc := func(wf *file_model.WeblensFileImpl) error {
		if wf.IsDir() {
			folderCount++
			return nil
		}
		index := strings.LastIndex(wf.Filename(), ".")
		size := wf.Size()
		if index == -1 {
			filetypeSizeMap["other"] += size
		} else {
			filetypeSizeMap[wf.Filename()[index+1:]] += size
		}

		return nil
	}

	err := meta.RootDir.RecursiveMap(sizeFunc)
	if err != nil {
		t.ReqNoErr(err)
	}

	returned := make([]extSize, 0, len(filetypeSizeMap))
	for name, value := range maps.All(filetypeSizeMap) {
		returned = append(returned, extSize{Name: name, Value: value})
	}

	// freeSpace := dataStore.GetFreeSpace(meta.rootDir.GetAbsPath())
	freeSpace := 0

	t.SetResult(task.TaskResult{"sizesByExtension": returned, "bytesFree": freeSpace})
	t.Success()
}

func HashFile(t *task.Task) {
	meta := t.GetMeta().(models.HashFileMeta)

	contentId, err := service.GenerateContentId(meta.File)
	t.ReqNoErr(err)

	if contentId == "" && meta.File.Size() != 0 {
		t.Fail(werror.ErrNoContentId)
	}

	t.Log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Hashed file [%s] to [%s]", meta.File.GetPortablePath(), contentId) })

	// TODO - sync database content id if this file is created before being added to db (i.e upload)
	// err = dataStore.SetContentId(meta.file, contentId)
	// if err != nil {
	// 	t.ErrorAndExit(err)
	// }

	t.SetResult(task.TaskResult{"contentId": contentId})

	poolStatus := t.GetTaskPool().Status()
	meta.Caster.PushTaskUpdate(
		t, client.TaskCompleteEvent, task.TaskResult{
			"filename":      meta.File.Filename(),
			"tasksTotal":    poolStatus.Total,
			"tasksComplete": poolStatus.Complete,
		},
	)

	t.Success()
}

func RegisterJobs(workerPool task.TaskService, role models.ServerRole) {
	log.Debug().Msgf("Registering jobs for %s", role)

	workerPool.RegisterJob(models.ScanDirectoryTask, ScanDirectory)
	workerPool.RegisterJob(models.ScanFileTask, ScanFile)
	workerPool.RegisterJob(models.UploadFilesTask, HandleFileUploads, task.TaskOptions{Persistent: true, Unique: true})
	workerPool.RegisterJob(models.CreateZipTask, CreateZip)
	workerPool.RegisterJob(models.GatherFsStatsTask, GatherFilesystemStats)
	workerPool.RegisterJob(models.BackupTask, DoBackup)
	workerPool.RegisterJob(models.CopyFileFromCoreTask, CopyFileFromCore)
	workerPool.RegisterJob(models.RestoreCoreTask, RestoreCore)
	workerPool.RegisterJob(models.HashFileTask, HashFile)
	// if role == models.BackupServerRole {
	// } else if role == models.CoreServerRole {
	// }
}
