package dataProcess

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/saracen/fastzip"
)

func ScanFile(meta ScanMetadata) {
	db := dataStore.NewDB(meta.Username)
	defer func(){meta.PartialMedia = nil}()
	if meta.PartialMedia == nil {
		m, _ := db.GetMediaByFile(meta.File, true)
		meta.PartialMedia = &m
	}
	err := ProcessMediaFile(meta.File, meta.PartialMedia, db)
	if err != nil {
		util.DisplayError(err, "Failed to process new meda file")
	}
}

func createZipFromPaths(t *task) {
	zipMeta := t.metadata.(ZipMetadata)

	if len(zipMeta.Files) == 0 {
		err := fmt.Errorf("cannot create a zip with no files")
		util.DisplayError(err)
		t.err = err

		return
	}

	filesInfoMap := map[string]os.FileInfo{}

	util.Map(zipMeta.Files,
		func (file *dataStore.WeblensFileDescriptor) error {
			file.RecursiveMap(func (f *dataStore.WeblensFileDescriptor) {
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
		util.DisplayError(err)
		t.err = err
		return
	}
	if zipExists {
		t.setResult(KeyVal{Key: "takeoutId", Val: zipFile.Id()})
		t.BroadcastComplete("zip_complete")
		return
	}

	fp, err := os.Create(zipFile.String())
	if err != nil {
		util.DisplayError(err)
		t.err = err
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
		util.DisplayError(err, "Archive Error")
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
		status := struct {CompletedFiles int `json:"completedFiles"`; TotalFiles int `json:"totalFiles"`; SpeedBytes int `json:"speedBytes"`} {CompletedFiles: int(entries), TotalFiles: totalFiles, SpeedBytes: int(bytes-prevBytes)}
		Broadcast("task", t.TaskId, "create_zip_progress", status)

		time.Sleep(time.Second)
	}
	if archiveErr != nil {
		t.err = *archiveErr
		util.DisplayError(*archiveErr, "Failed to archive")
		return
	}

	t.setResult(KeyVal{Key: "takeoutId", Val: zipFile.Id()})
	t.BroadcastComplete("zip_complete")
}

func moveFile(t *task) {
	moveMeta := t.metadata.(MoveMeta)

	file := dataStore.FsTreeGet(moveMeta.FileId)
	if file.Err() != nil {
		err := fmt.Errorf("could not find existing file")
		panic(err)
	}

	destinationFolder := dataStore.FsTreeGet(moveMeta.DestinationFolderId)
	preUpdateFile := file.Copy()
	err := dataStore.FsTreeMove(file, destinationFolder, moveMeta.NewFilename, false)
	if err != nil {
		util.DisplayError(err)
		t.err = err
		return
	}

	PushItemUpdate(preUpdateFile, file)

}