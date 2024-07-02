package dataProcess

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/dataStore/media"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/saracen/fastzip"
)

func scanFile(t *task) {
	meta := t.metadata.(scanMetadata)
	// util.Debug.Println("Starting scan file task", meta.file.GetAbsPath())

	if !meta.file.IsDisplayable() {
		t.ErrorAndExit(ErrNonDisplayable)
	}

	contentId := meta.file.GetContentId()
	if contentId == "" {
		t.ErrorAndExit(fmt.Errorf("trying to scan file with no content id: %s", meta.file.GetAbsPath()))
	}

	meta.partialMedia = media.New(contentId)
	if slices.ContainsFunc(
		meta.partialMedia.GetFiles(), func(fId types.FileId) bool {
			return fId == meta.file.ID()
		},
	) {
		t.success("Media already imported")
	}
	// if meta.partialMedia.IsImported() {
	// 	if meta.file.Owner() == meta.partialMedia.GetOwner() {
	// 		meta.partialMedia.AddFile(meta.file)
	// 		t.success("Media already imported")
	// 		return
	// 	}
	// }

	t.CheckExit()

	t.SetErrorCleanup(
		func() {
			meta.partialMedia.Clean()
		},
	)

	t.metadata = meta

	t.CheckExit()
	processMediaFile(t)
}

func createZipFromPaths(t *task) {
	zipMeta := t.metadata.(zipMetadata)

	if len(zipMeta.files) == 0 {
		t.ErrorAndExit(ErrEmptyZip)
	}

	filesInfoMap := map[string]os.FileInfo{}

	util.Map(
		zipMeta.files,
		func(file types.WeblensFile) error {
			return file.RecursiveMap(
				func(f types.WeblensFile) error {
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

	paths := util.MapToKeys(filesInfoMap)
	slices.Sort(paths)
	takeoutHash := util.GlobbyHash(8, strings.Join(paths, ""))
	zipFile, zipExists, err := dataStore.NewTakeoutZip(takeoutHash, zipMeta.username)
	if err != nil {
		t.ErrorAndExit(err)
	}
	if zipExists {
		t.setResult(types.TaskResult{"takeoutId": zipFile.ID().String()})
		t.caster.PushTaskUpdate(t, TaskCompleteEvent, t.result) // Let any client subscribers know we are done
		t.success()
		return
	}

	t.caster.PushTaskUpdate(t, TaskCreatedEvent, types.TaskResult{"totalFiles": len(filesInfoMap)})

	if zipMeta.shareId != "" {
		sh := types.SERV.ShareService.Get(zipMeta.shareId)
		if sh == nil {
			t.ErrorAndExit(types.ErrNoShare)
		}
		err := zipFile.SetShare(sh)
		if err != nil {
			t.ErrorAndExit(err)
		}
	}

	fp, err := os.Create(zipFile.GetAbsPath())
	if err != nil {
		t.ErrorAndExit(err)
	}
	defer func(fp *os.File) {
		err := fp.Close()
		if err != nil {
			util.ShowErr(err)
		}
	}(fp)

	a, err := fastzip.NewArchiver(
		fp, zipMeta.files[0].GetParent().GetAbsPath(), fastzip.WithStageDirectory(zipFile.GetParent().GetAbsPath()),
		fastzip.WithArchiverBufferSize(32),
	)

	if err != nil {
		t.ErrorAndExit(err)
	}

	defer func(a *fastzip.Archiver) {
		err := a.Close()
		if err != nil {
			util.ShowErr(err)
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

	bytesTotal := util.Reduce(
		zipMeta.files, func(file types.WeblensFile, acc int64) int64 {
			num, err := file.Size()
			if err != nil {
				util.ShowErr(err)
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
		util.Debug.Println(bytes, entries)
		if bytes != prevBytes {
			byteDiff := bytes - prevBytes
			timeNs := updateInterval * sinceUpdate

			t.caster.PushTaskUpdate(
				t, ZipProgressEvent, types.TaskResult{
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

	t.setResult(types.TaskResult{"takeoutId": zipFile.ID()})
	t.caster.PushTaskUpdate(t, ZipCompleteEvent, t.result) // Let any client subscribers know we are done
	t.success()
}

func moveFile(t *task) {
	moveMeta := t.metadata.(moveMeta)

	file := types.SERV.FileTree.Get(moveMeta.fileId)
	if file == nil {
		t.ErrorAndExit(errors.New("could not find existing file"))
	}

	acc := dataStore.NewAccessMeta(dataStore.WeblensRootUser)

	destinationFolder := types.SERV.FileTree.Get(moveMeta.destinationFolderId)
	if destinationFolder == destinationFolder.Owner().GetTrashFolder() {
		err := dataStore.MoveFileToTrash(file, acc, t.caster)
		if err != nil {
			t.ErrorAndExit(err, "Failed while assuming move file was to trash")
		}
		return
	} else if dataStore.IsFileInTrash(file) {
		err := dataStore.ReturnFileFromTrash(file, t.caster)
		if err != nil {
			t.ErrorAndExit(err, "Failed while assuming move file was out of trash")
		}
		return
	}
	err := types.SERV.FileTree.Move(
		file, destinationFolder, moveMeta.newFilename, false, t.caster.(types.BufferedBroadcasterAgent),
	)
	if err != nil {
		t.ErrorAndExit(err)
	}

	moveMeta.fileEvent.NewMoveAction(moveMeta.fileId, file.ID())
	t.success()
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

// Task for reading file chunks coming in from client requests, and writing them out
// to their corresponding files. Intention behind this implementation is to rid the
// client of interacting with any blocking calls to make the upload process as fast as
// possible, hopefully as fast as the slower of the 2 network speeds. This task handles
// everything *after* the client has had its data read into memory, this is the "bottom half"
// of the upload
func handleFileUploads(t *task) {
	meta := t.metadata.(writeFileMeta)

	t.CheckExit()

	rootFile := types.SERV.FileTree.Get(meta.rootFolderId)
	if rootFile == nil {
		t.ErrorAndExit(dataStore.ErrNoFile, "could not find root folder in upload. ID:", meta.rootFolderId)
	}

	var bottom, top, total int64
	// var bottom int
	var err error

	// This map will only be accessed by this task and by this 1 thread,
	// so we do not need any synchronization here
	fileMap := map[types.FileId]*fileUploadProgress{}

	var bufCaster types.BufferedBroadcasterAgent
	switch t.caster.(type) {
	case types.BufferedBroadcasterAgent:
		bufCaster = t.caster.(types.BufferedBroadcasterAgent)
	default:
		t.ErrorAndExit(ErrBadCaster)
	}

	bufCaster.DisableAutoFlush()
	var usingFiles []types.FileId
	var topLevels []types.WeblensFile

	// Release all the files once we are finished here, if they haven't been already.
	// This should only be required in error cases, as if all files are successfully
	// written, they are then unlocked in the main body.
	defer func() {
		for _, fId := range usingFiles {
			f := types.SERV.FileTree.Get(fId)
			if f != nil {
				err = f.RemoveTask(t.TaskId())
				if err != nil {
					util.ShowErr(err)
				}
			}
		}
	}()

WriterLoop:
	for {
		t.setTimeout(time.Now().Add(time.Second * 10))
		select {
		case signal := <-t.signalChan: // Listen for cancellation
			if signal == 1 {
				return
			}
		case chunk := <-meta.chunkStream:
			t.ClearTimeout()

			bottom, top, total, err = parseRangeHeader(chunk.ContentRange)
			if err != nil {
				t.ErrorAndExit(err)
			}

			if chunk.newFile != nil {

				tmpFile := chunk.newFile
				for tmpFile.GetParent() != rootFile {
					tmpFile = tmpFile.GetParent()
				}
				if tmpFile.GetParent() == rootFile && !slices.ContainsFunc(
					topLevels, func(f types.WeblensFile) bool { return f.ID() == tmpFile.ID() },
				) {
					topLevels = append(topLevels, tmpFile)
				}

				fileMap[chunk.newFile.ID()] = &fileUploadProgress{
					file: chunk.newFile, bytesWritten: 0, fileSizeTotal: total,
				}
				chunk.newFile.AddTask(t)
				util.InsertFunc(
					usingFiles, chunk.newFile.ID(),
					func(a, b types.FileId) int { return strings.Compare(a.String(), b.String()) },
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
			fileMap[chunk.FileId].bytesWritten += (top - bottom) + 1

			// Write the bytes to the real file
			err = fileMap[chunk.FileId].file.WriteAt(chunk.Chunk, bottom)
			if err != nil {
				util.ShowErr(err)
			}

			// When file is finished writing
			if fileMap[chunk.FileId].bytesWritten >= fileMap[chunk.FileId].fileSizeTotal {
				// Hash file content to get content ID. Must do this before attaching the file,
				// or the journal worker will beat us to it, which could break if scanning
				// the file shortly after uploading.
				_, err := dataStore.GenerateContentId(fileMap[chunk.FileId].file)
				if err != nil {
					t.ErrorAndExit(
						err, "failed generating content id for file", fileMap[chunk.FileId].file.GetAbsPath(),
					)
				}

				// Move the file from /tmp to its permanent location
				err = types.SERV.FileTree.AttachFile(fileMap[chunk.FileId].file, bufCaster)
				if err != nil {
					util.ShowErr(err)
				}

				// Unlock the file
				err = fileMap[chunk.FileId].file.RemoveTask(t.TaskId())
				if err != nil {
					util.ShowErr(err)
				}

				// Remove the file from our local map
				i, e := slices.BinarySearchFunc(
					usingFiles, chunk.FileId,
					func(a, b types.FileId) int { return strings.Compare(a.String(), b.String()) },
				)
				if e {
					util.Banish(usingFiles, i)
				}
				delete(fileMap, chunk.FileId)
			}

			// When we have no files being worked on, and there are no more
			// chunks to write, we are finished.
			if len(fileMap) == 0 && len(meta.chunkStream) == 0 {
				break WriterLoop
			}
			t.CheckExit()
			continue WriterLoop
		}
	}

	t.CheckExit()

	bufCaster.Flush()

	doingRootScan := false

	// Do not report that this task pool was created by this task, we want to detach
	// and allow these scans to take place independently
	newTp := types.SERV.WorkerPool.NewTaskPool(false, nil)
	for _, tl := range topLevels {
		if tl.IsDir() {
			bufCaster.PushFileUpdate(tl)
			newTp.ScanDirectory(tl, t.caster)
		} else if !doingRootScan {
			newTp.ScanDirectory(rootFile, t.caster)
			doingRootScan = true
		}
	}
	newTp.SignalAllQueued()

	if newTp.Status().Total != 0 {
		bufCaster.AutoFlushEnable()
		newTp.AddCleanup(
			func() {
				bufCaster.Close()
			},
		)
	} else {
		bufCaster.Close()
	}

	t.success()
}

func (t *task) NewFileInStream(file types.WeblensFile, fileSize int64) error {
	switch t.taskType {
	case WriteFileTask:
	default:
		return ErrBadTaskType
	}
	t.metadata.(writeFileMeta).chunkStream <- fileChunk{
		newFile: file, ContentRange: "0-0/" + strconv.FormatInt(fileSize, 10),
	}

	// We don't queue the upload task right away, we wait for the first file,
	// then we add the task to the queue here
	if t.queueState == PreQueued {
		t.Q(t.taskPool)
	}

	return nil
}

func (t *task) AddChunkToStream(fileId types.FileId, chunk []byte, contentRange string) error {
	switch t.metadata.(type) {
	case writeFileMeta:
	default:
		return ErrBadTaskType
	}
	chunkData := fileChunk{FileId: fileId, Chunk: chunk, ContentRange: contentRange}
	t.metadata.(writeFileMeta).chunkStream <- chunkData

	return nil
}

type extSize struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

func gatherFilesystemStats(t *task) {
	meta := t.metadata.(fsStatMeta)

	filetypeSizeMap := map[string]int64{}
	folderCount := 0

	// media := dataStore.GetMediaDir()
	// external := dataStore.GetExternalDir()
	// dataStore.ResizeDown(media)

	sizeFunc := func(wf types.WeblensFile) error {
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

	err := meta.rootDir.RecursiveMap(sizeFunc)
	if err != nil {
		t.ErrorAndExit(err)
	}

	ret := util.MapToSliceMutate(
		filetypeSizeMap, func(name string, value int64) extSize { return extSize{Name: name, Value: value} },
	)

	freeSpace := dataStore.GetFreeSpace(meta.rootDir.GetAbsPath())

	t.setResult(types.TaskResult{"sizesByExtension": ret, "bytesFree": freeSpace})
	t.success()
}

func hashFile(t *task) {
	meta := t.metadata.(hashFileMeta)

	if meta.file.IsDir() {
		t.ErrorAndExit(
			types.NewWeblensError("cannot hash directory"),
			meta.file.GetAbsPath(),
		)
	}

	if meta.file.GetContentId() != "" {
		t.success("Skipping file which already has content ID", meta.file.GetAbsPath())
	}

	fileSize, err := meta.file.Size()
	if err != nil {
		t.ErrorAndExit(err)
	}

	if fileSize == 0 {
		t.success("Skipping file with no content: ", meta.file.GetAbsPath())
		return
	}

	var contentId types.ContentId
	fp, err := meta.file.Read()
	if err != nil {
		t.ErrorAndExit(err)
	}
	defer func(fp *os.File) {
		err := fp.Close()
		if err != nil {
			util.ShowErr(err)
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
	contentId = types.ContentId(base64.URLEncoding.EncodeToString(newHash.Sum(nil)))[:20]
	// meta.file.SetContentId(contentId)
	t.setResult(types.TaskResult{"contentId": contentId})

	// TODO - sync database content id if this file is created before being added to db (i.e upload)
	// err = dataStore.SetContentId(meta.file, contentId)
	// if err != nil {
	// 	t.ErrorAndExit(err)
	// }

	t.success()
}
