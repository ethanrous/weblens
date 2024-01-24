package dataProcess

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/saracen/fastzip"
)

func ScanFile(t *Task) {
	meta := t.metadata.(ScanMetadata)

	displayable, err := meta.File.IsDisplayable()
	if err != nil && !errors.Is(err, dataStore.ErrNoMedia) {
		t.error(err)
	}
	if !displayable {
		err = errors.New(fmt.Sprint("attempt to process non-displayable file", meta.File.String()))
		t.error(err)
	}

	if meta.PartialMedia == nil {
		m, err := meta.File.GetMedia()
		if err != nil {
			t.error(err)
			return
		}
		meta.PartialMedia = m
	}

	t.SetErrorCleanup(func() {
		media, err := meta.File.GetMedia()
		util.DisplayError(err)
		if media != nil {
			media.Clean()
		}

		meta.File.ClearMedia()
	})

	processMediaFile(t)
}

func createZipFromPaths(t *Task) {
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

	takeoutHash := util.HashOfString(8, strings.Join(util.MapToKeys(filesInfoMap), ""))
	zipFile, zipExists, err := dataStore.NewTakeoutZip(takeoutHash)
	if err != nil {
		t.error(err)
		return
	}
	if zipExists {
		t.setResult(KeyVal{Key: "takeoutId", Val: zipFile.Id()})
		caster.PushTaskUpdate(t.taskId, "zip_complete", t.result) // Let any client subscribers know we are done
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
	totalFiles := len(filesInfoMap)

	// Update client over websocket until entire archive has been written, or an error is thrown
	for int64(totalFiles) > entries {
		if archiveErr != nil {
			break
		}
		bytes, entries = a.Written()
		prevBytes = bytes
		status := struct {
			CompletedFiles int `json:"completedFiles"`
			TotalFiles     int `json:"totalFiles"`
			SpeedBytes     int `json:"speedBytes"`
		}{CompletedFiles: int(entries), TotalFiles: totalFiles, SpeedBytes: int(bytes - prevBytes)}
		caster.PushTaskUpdate(t.taskId, "create_zip_progress", status)

		time.Sleep(time.Second)
	}
	if archiveErr != nil {
		t.error(*archiveErr)
		return
	}

	t.setResult(KeyVal{Key: "takeoutId", Val: zipFile.Id()})
	caster.PushTaskUpdate(t.taskId, "zip_complete", t.result) // Let any client subscribers know we are done
	t.success()
}

func moveFile(t *Task) {
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

func preloadThumbss(t *Task) {
	meta := t.metadata.(PreloadMetaMeta)
	paths := util.Map(meta.Files, func(f *dataStore.WeblensFile) string { return f.String() })

	et, err := exiftool.NewExiftool(exiftool.Api("largefilesupport"))
	if err != nil {
		util.DisplayError(err)
		t.err = err
	}

	et.ExtractMetadata(paths...)

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

type bufferChunk struct {
	lowByte    int
	highByte   int
	chunkBytes []byte
}

func writeToFile(t *Task) {
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

	file.AddTask(t)

	var buffer []bufferChunk
	var nextByte, min, max, total int

WriterLoop:
	for {
		t.setTimeout(time.Now().Add(time.Second * 60))
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

			chunk64 := string(chunk.Chunk)
			index := strings.Index(chunk64, ",")
			chunk64 = chunk64[index+1:]

			fileBytes, err := base64.StdEncoding.DecodeString(chunk64)
			if err != nil {
				t.error(err)
				return
			}

			currentBuf := bufferChunk{lowByte: min, highByte: max, chunkBytes: fileBytes}

			if nextByte != currentBuf.lowByte {
				util.Debug.Println("Buffering", currentBuf.lowByte, "-", currentBuf.highByte)
				buffer = append(buffer, currentBuf)
				continue WriterLoop
			}

			for {
				err = file.Append(currentBuf.chunkBytes)
				if err != nil {
					t.error(err)
					return
				}
				nextByte = currentBuf.highByte + 1

				if len(buffer) == 0 {
					break
				}

				var exists bool
				buffer, currentBuf, exists = util.YoinkFunc(buffer, func(b bufferChunk) bool { return b.lowByte == nextByte })
				if !exists {
					break
				}
			}

			if nextByte == total {
				break WriterLoop
			}
		}
	}

	caster.PushFileCreate(file)
	dataStore.Resize(parent)
	file.RemoveTask(t.TaskId())
	t.success()
}
