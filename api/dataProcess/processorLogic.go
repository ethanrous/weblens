package dataProcess

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/saracen/fastzip"
)

func ScanFile(meta ScanMetadata) {
	db := dataStore.NewDB(meta.Username)
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
	files := zipMeta.Files
	filePaths := util.Map(files, func (file *dataStore.WeblensFileDescriptor) string {return file.String()})

	filesInfoMap := map[string]os.FileInfo{}
	util.Map(files,
		func (file *dataStore.WeblensFileDescriptor) error {
			err := filepath.Walk(file.String(), func(pathname string, info os.FileInfo, err error) error {
				filesInfoMap[pathname] = info
				return nil
			})
			util.FailOnError(err, "Failed to walk directory selecting files to zip")
			return nil
		},
	)

	mapBytes, err := json.Marshal(filesInfoMap)
	util.FailOnError(err, "Failed to marshal zip files map")
	takeoutHash := util.HashOfString(8, string(mapBytes))
	t.setResult(KeyVal{Key: "takeoutId", Val: takeoutHash})

	zipPath := filepath.Join(util.GetTakeoutDir(), takeoutHash + ".zip")
	_, err = os.Stat(zipPath)
	if !errors.Is(err, fs.ErrNotExist) { // If the zip file already exists, then we're done
		t.setComplete("task", "zip_complete")
		return
	}

	zippy, err := os.Create(zipPath)
	util.FailOnError(err, "Could not create zip takeout file")

	chroot := filepath.Dir(filePaths[0])
	a, err := fastzip.NewArchiver(zippy, chroot, fastzip.WithStageDirectory(util.GetTakeoutDir()), fastzip.WithArchiverBufferSize(32))
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
		util.FailOnError(*archiveErr, "Failed to archive")
	}
	t.setComplete("task", "zip_complete")
}

func moveFile(t *task) {
	moveMeta := t.metadata.(MoveMeta)



	currentFile := dataStore.GetWFD(moveMeta.ParentFolderId, moveMeta.OldFilename)
	if currentFile.Err() != nil {
		err := fmt.Errorf("could not find existing file")
		panic(err)
	}

	opts := dataStore.CreateOpts().SetIgnoreNonexistance(true)
	destinationFile := dataStore.GetWFD(moveMeta.DestinationFolderId, moveMeta.NewFilename, opts)
	util.FailOnError(destinationFile.Err(), "")

	if destinationFile.Exists() {
		err := fmt.Errorf("destination file already exists")
		panic(err)
	}

	currentFile.Id()
	preUpdateFile := currentFile.Copy()
	if currentFile.Err() != nil {
		panic(currentFile.Err())
	}

	err := currentFile.MoveTo(destinationFile,
		func(taskType string, taskMeta map[string]any) {
			RequestTask(taskType, "", ScanMetadata{
					File: taskMeta["file"].(*dataStore.WeblensFileDescriptor),
					Username: taskMeta["username"].(string),
				})})
	if err != nil {
		panic(err)
	}

	PushItemUpdate(preUpdateFile, currentFile)

}