package jobs

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"maps"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/comm"
	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/models/service"
	"github.com/ethrousseau/weblens/task"
	"github.com/saracen/fastzip"
)

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
		}
		if tp != nil && tp.CreatedInTask() != nil {
			result["task_job_target"] = tp.CreatedInTask().GetMeta().(models.ScanMeta).File.Filename()
		} else if tp == nil {
			result["task_job_target"] = meta.File.Filename()
		}
	}

	if tp != nil {
		status := tp.Status()
		result["percent_progress"] = status.Progress
		result["tasks_complete"] = status.Complete
		result["tasks_total"] = status.Total
		result["runtime"] = status.Runtime
		if tp.CreatedInTask() != nil {
			result["task_job_name"] = tp.CreatedInTask().JobName()
		}
	} else {
		result["task_job_name"] = t.JobName()
	}

	return result
}

func CreateZip(t *task.Task) {
	zipMeta := t.GetMeta().(models.ZipMeta)

	if len(zipMeta.Files) == 0 {
		t.ErrorAndExit(werror.ErrEmptyZip)
	}

	filesInfoMap := map[string]os.FileInfo{}

	internal.Map(
		zipMeta.Files,
		func(file *fileTree.WeblensFile) error {
			return file.RecursiveMap(
				func(f *fileTree.WeblensFile) error {
					stat, err := os.Stat(f.GetAbsPath())
					if err != nil {
						t.ErrorAndExit(err)
					}
					filesInfoMap[f.GetAbsPath()] = stat
					return nil
				},
			)
		},
	)

	paths := slices.Sorted(maps.Keys(filesInfoMap))

	var takeoutKey string
	if len(zipMeta.Files) == 1 {
		takeoutKey = zipMeta.Files[0].Filename()
	} else {
		takeoutKey = internal.GlobbyHash(8, strings.Join(paths, ""))
	}

	zipName := takeoutKey
	var zipExists bool

	if !strings.HasSuffix(zipName, ".zip") {
		zipName = zipName + ".zip"
	}

	zipFile, err := zipMeta.FileService.NewZip(zipName, zipMeta.Requester)
	if err != nil && strings.Contains(err.Error(), "file already exists") {
		err = nil
		zipExists = true
	}

	if err != nil {

		t.ErrorAndExit(err)
	}

	if zipExists {
		t.SetResult(task.TaskResult{"takeoutId": zipFile.ID(), "filename": zipFile.Filename()})
		// Let any client subscribers know we are done
		zipMeta.Caster.PushTaskUpdate(t, comm.ZipCompleteEvent, t.GetResults())
		t.Success()
		return
	}

	zipMeta.Caster.PushTaskUpdate(t, comm.TaskCreatedEvent, task.TaskResult{"totalFiles": len(filesInfoMap)})

	// if zipMeta.Share != nil {
	// 	sh := types.SERV.ShareService.Get(zipMeta.shareId)
	// 	if sh == nil {
	// 		t.ErrorAndExit(types.ErrNoShare)
	// 	}
	// 	err := zipFile.SetShare(sh)
	// 	if err != nil {
	// 		t.ErrorAndExit(err)
	// 	}
	// }

	fp, err := os.Create(zipFile.GetAbsPath())
	if err != nil {
		t.ErrorAndExit(err)
	}
	defer func(fp *os.File) {
		err := fp.Close()
		if err != nil {
			log.ShowErr(err)
		}
	}(fp)

	a, err := fastzip.NewArchiver(
		fp, zipMeta.Files[0].GetParent().GetAbsPath(), fastzip.WithStageDirectory(zipFile.GetParent().GetAbsPath()),
		fastzip.WithArchiverBufferSize(32),
	)

	if err != nil {
		t.ErrorAndExit(err)
	}

	defer func(a *fastzip.Archiver) {
		err := a.Close()
		if err != nil {
			log.ShowErr(err)
		}
	}(a)

	var archiveErr *error

	// Shove archive to child thread so we can send updates with main thread
	go func() {
		err := a.Archive(context.Background(), filesInfoMap)
		if err != nil {
			archiveErr = &err
		}
	}()

	bytesTotal := internal.Reduce(
		zipMeta.Files, func(file *fileTree.WeblensFile, acc int64) int64 {
			num, err := file.Size()
			if err != nil {
				log.ShowErr(err)
			}
			return acc + num
		}, 0,
	)

	var entries int64
	var bytes int64
	var prevBytes int64 = -1
	var sinceUpdate int64 = 0
	totalFiles := len(filesInfoMap)

	const updateInterval = 500 * int64(time.Millisecond)

	// Update client over websocket until entire archive has been written, or an error is thrown
	for int64(totalFiles) > entries {
		if archiveErr != nil {
			break
		}
		sinceUpdate++
		bytes, entries = a.Written()
		if bytes != prevBytes {
			byteDiff := bytes - prevBytes
			timeNs := updateInterval * sinceUpdate

			zipMeta.Caster.PushTaskUpdate(
				t, comm.ZipProgressEvent, task.TaskResult{
					"completedFiles": int(entries), "totalFiles": totalFiles,
					"bytesSoFar": bytes,
					"bytesTotal": bytesTotal,
					"speedBytes": int((float64(byteDiff) / float64(timeNs)) * float64(time.Second)),
				},
			)
			prevBytes = bytes
			sinceUpdate = 0
		}

		time.Sleep(time.Duration(updateInterval))
	}
	if archiveErr != nil {
		t.ErrorAndExit(*archiveErr)
	}

	t.SetResult(task.TaskResult{"takeoutId": zipFile.ID(), "filename": zipFile.Filename()})
	zipMeta.Caster.PushTaskUpdate(
		t, comm.ZipCompleteEvent, t.GetResults(),
	) // Let any client subscribers know we are done
	t.Success()
}

func MoveFile(t *task.Task) {
	moveMeta := t.GetMeta().(models.MoveMeta)

	file, err := moveMeta.FileService.GetFileSafe(moveMeta.FileId, moveMeta.User, nil)
	if err != nil {
		t.ErrorAndExit(err)
	}

	destinationFolder, err := moveMeta.FileService.GetFileSafe(moveMeta.DestinationFolderId, moveMeta.User, nil)
	if err != nil {
		t.ErrorAndExit(err)
	}
	if moveMeta.FileService.IsFileInTrash(destinationFolder) {
		err = moveMeta.FileService.MoveFileToTrash(file, moveMeta.User, nil, moveMeta.Caster)
		if err != nil {
			t.ErrorAndExit(err, "Failed while assuming move file was into trash")
		}
		return
	} else if moveMeta.FileService.IsFileInTrash(file) {
		err = moveMeta.FileService.ReturnFilesFromTrash([]*fileTree.WeblensFile{file}, moveMeta.Caster)
		if err != nil {
			t.ErrorAndExit(err, "Failed while assuming move file was returning from trash")
		}
		return
	}

	err = moveMeta.FileService.MoveFile(
		file, destinationFolder, moveMeta.NewFilename,
		moveMeta.Caster,
	)
	if err != nil {
		t.ErrorAndExit(err)
	}

	// err = t.taskPool.workerPool.fileTree.GetJournal().LogEvent(moveEvent)
	// if err != nil {
	// 	t.ErrorAndExit(err)
	// }

	t.Success()
}

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
		t.ErrorAndExit(
			err, "could not get root folder in upload. ID:", meta.RootFolderId,
		)
	}

	var bottom, top, total int64

	// This map will only be accessed by this task and by this 1 thread,
	// so we do not need any synchronization here
	fileMap := map[fileTree.FileId]*models.FileUploadProgress{}

	// meta.Caster.DisableAutoFlush()
	var usingFiles []fileTree.FileId
	var topLevels []*fileTree.WeblensFile

	// TODO
	// Release all the files once we are finished here, if they haven't been already.
	// This should only be required in error cases, as if all files are successfully
	// written, they are then unlocked in the main body.
	// defer func() {
	// 	for _, fId := range usingFiles {
	// 		f := t.taskPool.workerPool.fileTree.Get(fId)
	// 		if f != nil {
	// 			err = f.RemoveTask(t.TaskId())
	// 			if err != nil {
	// 				wlog.ShowErr(err)
	// 			}
	// 		}
	// 	}
	// }()

	fileEvent := meta.FileService.GetMediaJournal().NewEvent()
	defer func() {
		if !t.CheckExit() {
			meta.FileService.GetMediaJournal().LogEvent(fileEvent)
		}
	}()

WriterLoop:
	for {
		t.SetTimeout(time.Now().Add(time.Second * 10))
		select {
		case signal := <-t.GetSignalChan(): // Listen for cancellation
			if signal == 1 {
				return
			}
		case chunk := <-meta.ChunkStream:
			t.ClearTimeout()

			bottom, top, total, err = parseRangeHeader(chunk.ContentRange)
			if err != nil {
				t.ErrorAndExit(err)
			}

			if chunk.NewFile != nil {

				tmpFile := chunk.NewFile
				for tmpFile.GetParent() != rootFile {
					tmpFile = tmpFile.GetParent()
				}
				if tmpFile.GetParent() == rootFile && !slices.ContainsFunc(
					topLevels, func(f *fileTree.WeblensFile) bool { return f.ID() == tmpFile.ID() },
				) {
					topLevels = append(topLevels, tmpFile)
				}

				fileMap[chunk.NewFile.ID()] = &models.FileUploadProgress{
					File: chunk.NewFile, BytesWritten: 0, FileSizeTotal: total,
				}

				internal.InsertFunc(
					usingFiles, chunk.NewFile.ID(),
					func(a, b fileTree.FileId) int { return strings.Compare(string(a), string(b)) },
				)
				continue WriterLoop
			}

			// We use `0-0/-1` as a fake "range header" to indicate that the upload for
			// the specific file has had an error or been canceled, and should be removed.
			if total == -1 {
				delete(fileMap, chunk.FileId)
			}

			// Add the new bytes to the counter for the file-size of this file.
			// If we upload content in range e.g. 0-1 bytes, that includes 2 bytes,
			// but top - bottom (1 - 0) is 1, so we add 1 to match
			fileMap[chunk.FileId].BytesWritten += (top - bottom) + 1

			// Write the bytes to the real file
			err = fileMap[chunk.FileId].File.WriteAt(chunk.Chunk, bottom)
			if err != nil {
				log.ShowErr(err)
			}

			// When file is finished writing
			if fileMap[chunk.FileId].BytesWritten >= fileMap[chunk.FileId].FileSizeTotal {
				// Hash file content to get content ID. Must do this before attaching the file,
				// or the journal worker will beat us to it, which could break if scanning
				// the file shortly after uploading.
				_, err := service.GenerateContentId(fileMap[chunk.FileId].File)
				if err != nil {
					t.ErrorAndExit(
						err, "failed generating content id for file", fileMap[chunk.FileId].File.GetAbsPath(),
					)
				}

				// Move the file from /tmp to its permanent location
				// err = meta.FileService.AttachFile(fileMap[chunk.FileId].File, meta.Caster)
				// if err != nil {
				// 	wlog.ShowErr(err)
				// }

				fileEvent.NewCreateAction(fileMap[chunk.FileId].File)

				// Remove the file from our local map
				i, e := slices.BinarySearchFunc(
					usingFiles, chunk.FileId,
					func(a, b fileTree.FileId) int { return strings.Compare(string(a), string(b)) },
				)
				if e {
					internal.Banish(usingFiles, i)
				}
				delete(fileMap, chunk.FileId)
			}

			// When we have no files being worked on, and there are no more
			// chunks to write, we are finished.
			if len(fileMap) == 0 && len(meta.ChunkStream) == 0 {
				break WriterLoop
			}
			t.ExitIfSignaled()
			continue WriterLoop
		}
	}

	t.ExitIfSignaled()

	doingRootScan := false

	// Do not report that this task pool was created by this task, we want to detach
	// and allow these scans to take place independently
	newTp := t.GetTaskPool().GetWorkerPool().NewTaskPool(false, nil)
	for _, tl := range topLevels {
		media := meta.MediaService.Get(models.ContentId(tl.GetContentId()))
		meta.Caster.PushFileUpdate(tl, media)
		if tl.IsDir() {
			// err = t.taskPool.workerPool.fileTree.ResizeDown(tl, t.caster)
			// if err != nil {
			// 	wlog.ShowErr(err)
			// }
			err = meta.FileService.ResizeDown(tl, meta.Caster)
			if err != nil {
				t.ErrorAndExit(err)
			}

			if err != nil {
				t.ErrorAndExit(err)
			}

			scanMeta := models.ScanMeta{
				File:         tl,
				FileService:  meta.FileService,
				TaskService:  meta.TaskService,
				MediaService: meta.MediaService,
				TaskSubber:   meta.TaskSubber,
				Caster:       meta.Caster,
			}
			_, err = t.GetTaskPool().GetWorkerPool().DispatchJob(models.ScanDirectoryTask, scanMeta, newTp)
			if err != nil {
				log.ErrTrace(err)
				continue
			}
		} else if !doingRootScan {
			scanMeta := models.ScanMeta{
				File:         rootFile,
				FileService:  meta.FileService,
				TaskService:  meta.TaskService,
				MediaService: meta.MediaService,
				TaskSubber:   meta.TaskSubber,
				Caster:       meta.Caster,
			}
			_, err = t.GetTaskPool().GetWorkerPool().DispatchJob(models.ScanDirectoryTask, scanMeta, newTp)
			if err != nil {
				log.ErrTrace(err)
				continue
			}
			doingRootScan = true
		}
	}
	newTp.SignalAllQueued()

	err = meta.FileService.ResizeUp(rootFile, meta.Caster)
	if err != nil {
		t.ErrorAndExit(err)
	}

	if newTp.Status().Total != 0 {
		newTp.AddCleanup(
			func(_ task.Pool) {
				meta.Caster.Close()
			},
		)
	} else {
		meta.Caster.Close()
	}

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

	sizeFunc := func(wf *fileTree.WeblensFile) error {
		if wf.IsDir() {
			folderCount++
			return nil
		}
		index := strings.LastIndex(wf.Filename(), ".")
		size, err := wf.Size()
		if err != nil {
			return err
		}
		if index == -1 {
			filetypeSizeMap["other"] += size
		} else {
			filetypeSizeMap[wf.Filename()[index+1:]] += size
		}

		return nil
	}

	err := meta.RootDir.RecursiveMap(sizeFunc)
	if err != nil {
		t.ErrorAndExit(err)
	}

	ret := internal.MapToSliceMutate(
		filetypeSizeMap, func(name string, value int64) extSize { return extSize{Name: name, Value: value} },
	)

	// freeSpace := dataStore.GetFreeSpace(meta.rootDir.GetAbsPath())
	freeSpace := 0

	t.SetResult(task.TaskResult{"sizesByExtension": ret, "bytesFree": freeSpace})
	t.Success()
}

func HashFile(t *task.Task) {
	meta := t.GetMeta().(models.HashFileMeta)

	if meta.File.IsDir() {
		t.ErrorAndExit(
			errors.New("cannot hash directory"),
			meta.File.GetAbsPath(),
		)
	}

	if meta.File.GetContentId() != "" {
		t.Success("Skipping file which already has content ID", meta.File.GetAbsPath())
	}

	fileSize, err := meta.File.Size()
	if err != nil {
		t.ErrorAndExit(err)
	}

	if fileSize == 0 {
		t.Success("Skipping file with no content: ", meta.File.GetAbsPath())
		return
	}

	var contentId models.ContentId
	fp, err := meta.File.Readable()
	if err != nil {
		t.ErrorAndExit(err)
	}
	defer func(fp *os.File) {
		err := fp.Close()
		if err != nil {
			log.ShowErr(err)
		}
	}(fp)

	// Read up to 1MB at a time
	bufSize := math.Min(float64(fileSize), 1000*1000)

	buf := make([]byte, int64(bufSize))

	newHash := sha256.New()
	_, err = io.CopyBuffer(newHash, fp, buf)
	if err != nil {
		t.ErrorAndExit(err)
	}
	contentId = models.ContentId(base64.URLEncoding.EncodeToString(newHash.Sum(nil)))[:20]
	// meta.file.SetContentId(contentId)
	t.SetResult(task.TaskResult{"contentId": contentId})

	// TODO - sync database content id if this file is created before being added to db (i.e upload)
	// err = dataStore.SetContentId(meta.file, contentId)
	// if err != nil {
	// 	t.ErrorAndExit(err)
	// }

	t.Success()
}
