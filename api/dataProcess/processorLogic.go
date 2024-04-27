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
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/saracen/fastzip"
)

func scanFile(t *task) {
	meta := t.metadata.(ScanMetadata)
	// util.Debug.Println("Starting scan file task", meta.file.GetAbsPath())

	if !meta.file.IsDisplayable() {
		t.ErrorAndExit(ErrNonDisplayable)
	}

	contentId := meta.file.GetContentId()
	if contentId == "" {
		t.ErrorAndExit(fmt.Errorf("trying to scan file with no content id: %s", meta.file.GetAbsPath()))
	}

	meta.partialMedia = dataStore.NewMedia(contentId)
	if meta.partialMedia.IsImported() && meta.file.Owner() == meta.partialMedia.GetOwner() {
		meta.partialMedia.AddFile(meta.file)
		t.success("Media already imported")
		return
	}

	t.CheckExit()

	t.SetErrorCleanup(func() {
		meta.partialMedia.Clean()
		// media, err := meta.file.GetMedia()
		// if err != nil && err != dataStore.ErrNoMedia {
		// 	util.ErrTrace(err)
		// }
		// if media != nil {
		// 	media.Clean()
		// }

		// meta.file.ClearMedia()
	})

	t.metadata = meta

	t.CheckExit()
	processMediaFile(t)
}

func createZipFromPaths(t *task) {
	zipMeta := t.metadata.(ZipMetadata)

	if len(zipMeta.files) == 0 {
		t.ErrorAndExit(ErrEmptyZip)
	}

	filesInfoMap := map[string]os.FileInfo{}

	util.Map(zipMeta.files,
		func(file types.WeblensFile) error {
			return file.RecursiveMap(func(f types.WeblensFile) error {
				stat, err := os.Stat(f.GetAbsPath())
				if err != nil {
					t.ErrorAndExit(err)
				}
				filesInfoMap[f.GetAbsPath()] = stat
				return nil
			})
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
		t.setResult(types.TaskResult{"takeoutId": zipFile.Id().String()})
		t.caster.PushTaskUpdate(t.taskId, TaskComplete, t.result) // Let any client subscribers know we are done
		t.success()
		return
	}

	if zipMeta.shareId != "" {
		s, err := dataStore.GetShare(zipMeta.shareId, dataStore.FileShare)
		if err != nil {
			t.ErrorAndExit(err)
		}
		zipFile.AppendShare(s)
	}

	fp, err := os.Create(zipFile.GetAbsPath())
	if err != nil {
		t.ErrorAndExit(err)
	}
	defer fp.Close()

	a, err := fastzip.NewArchiver(fp, zipMeta.files[0].GetParent().GetAbsPath(), fastzip.WithStageDirectory(zipFile.GetParent().GetAbsPath()), fastzip.WithArchiverBufferSize(32))
	util.FailOnError(err, "Filed to create new zip archiver")
	defer a.Close()

	var archiveErr *error

	// Shove archive to child thread so we can send updates with main thread
	go func() {
		err := a.Archive(context.Background(), filesInfoMap)
		if err != nil {
			archiveErr = &err
		}
	}()

	var entries int64
	var bytes int64
	var prevBytes int64 = -1
	var sinceUpdate int64 = 0
	totalFiles := len(filesInfoMap)

	const UPDATE_INTERVAL int64 = 500 * int64(time.Millisecond)

	// Update client over websocket until entire archive has been written, or an error is thrown
	for int64(totalFiles) > entries {
		if archiveErr != nil {
			break
		}
		sinceUpdate++
		bytes, entries = a.Written()
		if bytes != prevBytes {
			byteDiff := bytes - prevBytes
			timeNs := UPDATE_INTERVAL * sinceUpdate

			t.caster.PushTaskUpdate(t.taskId, TaskProgress, types.TaskResult{"completedFiles": int(entries), "totalFiles": totalFiles, "speedBytes": int((float64(byteDiff) / float64(timeNs)) * float64(time.Second))})
			prevBytes = bytes
			sinceUpdate = 0
		}

		time.Sleep(time.Duration(UPDATE_INTERVAL))
	}
	if archiveErr != nil {
		t.ErrorAndExit(*archiveErr)
	}

	t.setResult(types.TaskResult{"takeoutId": zipFile.Id()})
	t.caster.PushTaskUpdate(t.taskId, TaskComplete, t.result) // Let any client subscribers know we are done
	t.success()
}

func moveFile(t *task) {
	moveMeta := t.metadata.(MoveMeta)

	file := dataStore.FsTreeGet(moveMeta.fileId)
	if file == nil {
		t.ErrorAndExit(errors.New("could not find existing file"))
	}

	destinationFolder := dataStore.FsTreeGet(moveMeta.destinationFolderId)
	if destinationFolder == destinationFolder.Owner().GetTrashFolder() {
		err := dataStore.MoveFileToTrash(file, t.caster)
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
	err := dataStore.FsTreeMove(file, destinationFolder, moveMeta.newFilename, false, t.caster.(types.BufferedBroadcasterAgent))
	if err != nil {
		t.ErrorAndExit(err)
	}
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

func handleFileUploads(t *task) {
	meta := t.metadata.(WriteFileMeta)

	t.CheckExit()

	var min, max, total int64
	// var min int
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
	usingFiles := []types.FileId{}
	defer func() {
		for _, fId := range usingFiles {
			f := dataStore.FsTreeGet(fId)
			if f != nil {
				f.RemoveTask(t.TaskId())
			}
		}
	}()

WriterLoop:
	for {
		t.setTimeout(time.Now().Add(time.Second))
		select {
		case signal := <-t.signalChan: // Listen for cancellation
			if signal == 1 {
				return
			}
		case chunk := <-meta.chunkStream:
			t.ClearTimeout()

			min, max, total, err = parseRangeHeader(chunk.ContentRange)
			if err != nil {
				t.ErrorAndExit(err)
			}

			if chunk.newFile != nil {
				fileMap[chunk.newFile.Id()] = &fileUploadProgress{file: chunk.newFile, bytesWritten: 0, fileSizeTotal: total}
				chunk.newFile.AddTask(t)
				util.InsertFunc(usingFiles, chunk.newFile.Id(), func(a, b types.FileId) int { return strings.Compare(a.String(), b.String()) })
				continue WriterLoop
			}

			// We use `0-0/-1` as a fake "range header" to indicate that the upload for
			// the specific file has had an error or been canceled, and should be removed.
			if total == -1 {
				delete(fileMap, chunk.FileId)
			}

			fileMap[chunk.FileId].bytesWritten += (max - min) + 1

			fileMap[chunk.FileId].file.WriteAt(chunk.Chunk, int64(min))

			if fileMap[chunk.FileId].bytesWritten >= fileMap[chunk.FileId].fileSizeTotal {
				dataStore.AttachFile(fileMap[chunk.FileId].file, bufCaster)
				fileMap[chunk.FileId].file.RemoveTask(t.TaskId())
				i, e := slices.BinarySearchFunc(usingFiles, chunk.FileId, func(a, b types.FileId) int { return strings.Compare(a.String(), b.String()) })
				if e {
					util.Banish(usingFiles, i)
				}
				delete(fileMap, chunk.FileId)
			}
			if len(fileMap) == 0 && len(meta.chunkStream) == 0 {
				break WriterLoop
			}
			t.CheckExit()
			continue WriterLoop
		}
	}

	t.CheckExit()
	rootFile := dataStore.FsTreeGet(meta.rootFolderId)
	dataStore.ResizeDown(rootFile, bufCaster)
	dataStore.ResizeUp(rootFile, bufCaster)
	bufCaster.Close()
	t.success()
}

func (t *task) NewFileInStream(file types.WeblensFile, fileSize int64) error {
	switch t.metadata.(type) {
	case WriteFileMeta:
	default:
		return ErrBadTaskMetaType
	}
	t.metadata.(WriteFileMeta).chunkStream <- FileChunk{newFile: file, ContentRange: "0-0/" + strconv.FormatInt(fileSize, 10)}

	if t.taskPool == nil {
		GetGlobalQueue().QueueTask(t)
	}

	return nil
}

func (t *task) AddChunkToStream(fileId types.FileId, chunk []byte, contentRange string) error {
	switch t.metadata.(type) {
	case WriteFileMeta:
	default:
		return ErrBadTaskMetaType
	}
	chunkData := FileChunk{FileId: fileId, Chunk: chunk, ContentRange: contentRange}
	t.metadata.(WriteFileMeta).chunkStream <- chunkData

	return nil
}

type extSize struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

func gatherFilesystemStats(t *task) {
	meta := t.metadata.(FsStatMeta)

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

	ret := util.MapToSliceMutate(filetypeSizeMap, func(name string, value int64) extSize { return extSize{Name: name, Value: value} })

	freeSpace := dataStore.GetFreeSpace(meta.rootDir.GetAbsPath())

	t.setResult(types.TaskResult{"sizesByExtension": ret, "bytesFree": freeSpace})
	t.success()
}

func hashFile(t *task) {
	meta := t.metadata.(HashFileMeta)

	if meta.file.IsDir() {
		t.ErrorAndExit(dataStore.ErrDirNotAllowed, meta.file.GetAbsPath())
	}

	fileSize, err := meta.file.Size()
	if err != nil {
		t.ErrorAndExit(err)
	}

	if fileSize == 0 {
		t.success("Skipping file with no content", meta.file.GetAbsPath())
		return
	}

	var contentId types.ContentId
	fp, err := meta.file.Read()
	if err != nil {
		t.ErrorAndExit(err)
	}
	defer fp.Close()

	// Read up to 1MB at a time
	bufSize := math.Min(float64(fileSize), 1000*1000)

	buf := make([]byte, int64(bufSize))

	newHash := sha256.New()
	_, err = io.CopyBuffer(newHash, fp, buf)
	if err != nil {
		t.ErrorAndExit(err)
	}
	contentId = types.ContentId(base64.URLEncoding.EncodeToString(newHash.Sum(nil)))[:20]
	err = dataStore.SetContentId(meta.file, contentId)
	if err != nil {
		t.ErrorAndExit(err)
	}

	t.success()
}
