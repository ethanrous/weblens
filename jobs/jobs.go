package jobs

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/service"
	"github.com/ethanrous/weblens/task"
	"github.com/saracen/fastzip"
)

func CreateZip(t *task.Task) {
	zipMeta := t.GetMeta().(models.ZipMeta)

	if len(zipMeta.Files) == 0 {
		t.ReqNoErr(werror.ErrEmptyZip)
	}

	filesInfoMap := map[string]os.FileInfo{}

	internal.Map(
		zipMeta.Files,
		func(file *fileTree.WeblensFileImpl) error {
			return file.RecursiveMap(
				func(f *fileTree.WeblensFileImpl) error {
					stat, err := os.Stat(f.AbsPath())
					if err != nil {
						t.ReqNoErr(err)
					}
					filesInfoMap[f.AbsPath()] = stat
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
	} else if err != nil {
		t.ReqNoErr(err)
	}

	if zipExists {
		t.SetResult(task.TaskResult{"takeoutId": zipFile.ID(), "filename": zipFile.Filename()})
		// Let any client subscribers know we are done
		zipMeta.Caster.PushTaskUpdate(t, models.ZipCompleteEvent, t.GetResults())
		t.Success()
		return
	}

	zipMeta.Caster.PushTaskUpdate(t, models.TaskCreatedEvent, task.TaskResult{"totalFiles": len(filesInfoMap)})

	fp, err := os.Create(zipFile.AbsPath())
	if err != nil {
		t.ReqNoErr(err)
	}
	defer func(fp *os.File) {
		err := fp.Close()
		if err != nil {
			log.ShowErr(err)
		}
	}(fp)

	a, err := fastzip.NewArchiver(
		fp, zipMeta.Files[0].GetParent().AbsPath(),
		fastzip.WithStageDirectory(zipFile.GetParent().AbsPath()),
		fastzip.WithArchiverBufferSize(1024*1024*1024),
		fastzip.WithArchiverMethod(zip.Store),
	)

	if err != nil {
		t.ReqNoErr(err)
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
		zipMeta.Files, func(file *fileTree.WeblensFileImpl, acc int64) int64 {
			num := file.Size()
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
				t, models.ZipProgressEvent, task.TaskResult{
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
		t.ReqNoErr(*archiveErr)
	}

	t.SetResult(task.TaskResult{"takeoutId": zipFile.ID(), "filename": zipFile.Filename()})
	zipMeta.Caster.PushTaskUpdate(
		t, models.ZipCompleteEvent, t.GetResults(),
	) // Let any client subscribers know we are done
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
		t.ReqNoErr(err)
	}

	var bottom, top, total int64

	// This map will only be accessed by this task and by this 1 thread,
	// so we do not need any synchronization here
	fileMap := map[fileTree.FileId]*models.FileUploadProgress{}

	// meta.Caster.DisableAutoFlush()
	var usingFiles []fileTree.FileId
	var topLevels []*fileTree.WeblensFileImpl

	// TODO
	// Release all the files once we are finished here, if they haven't been already.
	// This should only be required in error cases, as if all files are successfully
	// written, they are then unlocked in the main body.

	fileEvent := meta.FileService.GetJournalByTree("USERS").NewEvent()
	defer func() {
		if !t.CheckExit() {
			meta.FileService.GetJournalByTree("USERS").LogEvent(fileEvent)
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
				t.ReqNoErr(err)
			}

			if chunk.NewFile != nil {

				tmpFile := chunk.NewFile
				for tmpFile.GetParent() != rootFile {
					tmpFile = tmpFile.GetParent()
				}
				if tmpFile.GetParent() == rootFile && !slices.ContainsFunc(
					topLevels, func(f *fileTree.WeblensFileImpl) bool { return f.ID() == tmpFile.ID() },
				) {
					topLevels = append(topLevels, tmpFile)
				}

				fileMap[chunk.NewFile.ID()] = &models.FileUploadProgress{
					File: chunk.NewFile, BytesWritten: 0, FileSizeTotal: total, Hash: sha256.New(),
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

			chnk := fileMap[chunk.FileId]

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

				if !chnk.File.IsDir() {
					meta.Caster.PushFileCreate(chnk.File)
				}

				fileEvent.NewCreateAction(chnk.File)

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
		if tl.IsDir() {
			err = meta.FileService.ResizeDown(tl, meta.Caster)
			if err != nil {
				t.ReqNoErr(err)
			}

			if err != nil {
				t.ReqNoErr(err)
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
		media := meta.MediaService.Get(tl.GetContentId())
		if tl.IsDir() {
			meta.Caster.PushFileUpdate(tl, media)
		}
	}
	newTp.SignalAllQueued()

	err = meta.FileService.ResizeUp(rootFile, meta.Caster)
	if err != nil {
		t.ReqNoErr(err)
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

	sizeFunc := func(wf *fileTree.WeblensFileImpl) error {
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

	contentId, err := service.GenerateContentId(meta.File)
	t.ReqNoErr(err)

	log.Trace.Printf("Hashed file %s to %s", meta.File.GetPortablePath(), contentId)

	// TODO - sync database content id if this file is created before being added to db (i.e upload)
	// err = dataStore.SetContentId(meta.file, contentId)
	// if err != nil {
	// 	t.ErrorAndExit(err)
	// }

	t.SetResult(task.TaskResult{"contentId": contentId})

	poolStatus := t.GetTaskPool().Status()
	meta.Caster.PushTaskUpdate(
		t, models.TaskCompleteEvent, task.TaskResult{
			"fileName":      meta.File.Filename(),
			"tasksTotal":    poolStatus.Total,
			"tasksComplete": poolStatus.Complete,
		},
	)

	t.Success()
}
