package dataProcess

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/saracen/fastzip"
)

func ScanFile(t *task) {
	meta := t.metadata.(ScanMetadata)

	displayable, err := meta.File.IsDisplayable()
	if err != nil && !errors.Is(err, dataStore.ErrNoMedia) {
		t.error(err)
	}
	if !displayable {
		err = errors.New(fmt.Sprintln("attempt to process non-displayable file", meta.File.String()))
		t.error(err)
		return
	}

	if meta.PartialMedia == nil {
		meta.PartialMedia = &dataStore.Media{}
	}

	t.SetErrorCleanup(func() {
		media, err := meta.File.GetMedia()
		if err != nil && err != dataStore.ErrNoMedia {
			util.DisplayError(err)
		}
		if media != nil {
			media.Clean()
		}

		meta.File.ClearMedia()
	})

	t.metadata = meta
	processMediaFile(t)
}

func createZipFromPaths(t *task) {
	zipMeta := t.metadata.(ZipMetadata)

	if len(zipMeta.Files) == 0 {
		t.error(errors.New("cannot create a zip with no files"))
		return
	}

	filesInfoMap := map[string]os.FileInfo{}

	util.Map(zipMeta.Files,
		func(file *dataStore.WeblensFile) error {
			file.RecursiveMap(func(f *dataStore.WeblensFile) {
				stat, err := os.Stat(f.String())
				util.FailOnError(err, "Failed to stat file %s", f.String())
				filesInfoMap[f.String()] = stat
			})
			return nil
		},
	)

	paths := util.MapToKeys(filesInfoMap)
	slices.Sort(paths)
	takeoutHash := util.GlobbyHash(8, strings.Join(paths, ""))
	zipFile, zipExists, err := dataStore.NewTakeoutZip(takeoutHash)
	if err != nil {
		t.error(err)
		return
	}
	if zipExists {
		t.setResult(KeyVal{Key: "takeoutId", Val: zipFile.Id()})
		t.caster.PushTaskUpdate(t.taskId, "zip_complete", t.result) // Let any client subscribers know we are done
		t.success()
		return
	}

	fp, err := os.Create(zipFile.String())
	if err != nil {
		t.error(err)
		return
	}
	defer fp.Close()

	a, err := fastzip.NewArchiver(fp, zipMeta.Files[0].GetParent().String(), fastzip.WithStageDirectory(zipFile.GetParent().String()), fastzip.WithArchiverBufferSize(32))
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
			status := struct {
				CompletedFiles int `json:"completedFiles"`
				TotalFiles     int `json:"totalFiles"`
				SpeedBytes     int `json:"speedBytes"`
			}{
				CompletedFiles: int(entries),
				TotalFiles:     totalFiles,
				SpeedBytes:     int((float64(byteDiff) / float64(timeNs)) * float64(time.Second)),
			}
			t.caster.PushTaskUpdate(t.taskId, "create_zip_progress", status)
			prevBytes = bytes
			sinceUpdate = 0
		}

		time.Sleep(time.Duration(UPDATE_INTERVAL))
	}
	if archiveErr != nil {
		t.error(*archiveErr)
		return
	}

	t.setResult(KeyVal{Key: "takeoutId", Val: zipFile.Id()})
	t.caster.PushTaskUpdate(t.taskId, "zip_complete", t.result) // Let any client subscribers know we are done
	t.success()
}

func moveFile(t *task) {
	moveMeta := t.metadata.(MoveMeta)

	file := dataStore.FsTreeGet(moveMeta.FileId)
	if file == nil {
		t.error(errors.New("could not find existing file"))
		return
	}

	destinationFolder := dataStore.FsTreeGet(moveMeta.DestinationFolderId)
	err := dataStore.FsTreeMove(file, destinationFolder, moveMeta.NewFilename, false)
	if err != nil {
		t.error(err)
		return
	}
}

func parseRangeHeader(contentRange string) (min, max, total int, err error) {
	rangeAndSize := strings.Split(contentRange, "/")
	rangeParts := strings.Split(rangeAndSize[0], "-")

	min, err = strconv.Atoi(rangeParts[0])
	if err != nil {
		return
	}

	max, err = strconv.Atoi(rangeParts[1])
	if err != nil {
		return
	}

	total, err = strconv.Atoi(rangeAndSize[1])
	if err != nil {
		return
	}
	return
}

func writeToFile(t *task) {
	meta := t.metadata.(WriteFileMeta)

	parent := dataStore.FsTreeGet(meta.ParentFolderId)
	if parent == nil {
		t.error(errors.New("failed to get parent of file to upload"))
		return
	}

	file, err := dataStore.Touch(parent, meta.Filename, true)
	if err != nil {
		t.error(err)
		return
	}

	// We have now really created the file, if this task fails, we will remove it
	t.SetErrorCleanup(func() {
		dataStore.FsTreeRemove(file)
	})

	file.AddTask(t)

	var min, max, total int

WriterLoop:
	for {
		t.setTimeout(time.Now().Add(time.Second * 30))
		select {
		case signal := <-t.signalChan: // Listen for cancellation
			if signal == 1 {
				return
			}
		case chunk := <-meta.ChunkStream:
			min, max, total, err = parseRangeHeader(chunk.ContentRange)
			if err != nil {
				t.error(err)
				return
			}

			file.WriteAt(chunk.Chunk, int64(min))

			// Uploading an entire 100 byte file would have the content range header
			// 0-99/100, so max is 99 and total is 100, so we -1.
			if max == total-1 {
				break WriterLoop
			}
			continue
		}
	}

	globalCaster.PushFileCreate(file)
	dataStore.Resize(parent)
	file.RemoveTask(t.TaskId())
	t.success()
}
