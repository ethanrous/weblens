package dataProcess

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/saracen/fastzip"
)

func ScanFile(meta ScanMetadata) {
	db := dataStore.NewDB("")
	if meta.PartialMedia == nil {
		m, err := meta.File.GetMedia()
		if err != nil {
			util.DisplayError(err)
			return
		}

		meta.PartialMedia = m
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
		status := struct {
			CompletedFiles int `json:"completedFiles"`
			TotalFiles     int `json:"totalFiles"`
			SpeedBytes     int `json:"speedBytes"`
		}{CompletedFiles: int(entries), TotalFiles: totalFiles, SpeedBytes: int(bytes - prevBytes)}
		caster.PushTaskUpdate(t.TaskId, "create_zip_progress", status)

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
	err := dataStore.FsTreeMove(file, destinationFolder, moveMeta.NewFilename, false)
	if err != nil {
		util.DisplayError(err)
		t.err = err
		return
	}

}

func preloadThumbs(t *task) {
	meta := t.metadata.(PreloadMetaMeta)
	paths := util.Map(meta.Files, func(f *dataStore.WeblensFile) string { return f.String() })

	if len(paths) == 0 {
		t.err = fmt.Errorf("failed to parse raw thumbnails, files slice is empty")
		util.DisplayError(t.err.(error))
		return
	}

	allPathsStr := strings.Join(util.Map(paths, func(path string) string { return fmt.Sprintf("\"%s\"", path) }), " ")
	cmdString := fmt.Sprintf("exiftool -a -b -%s %s", meta.ExifThumbType, allPathsStr)
	cmd := exec.Command("/bin/bash", "-c", cmdString)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		util.Warning.Printf("Failed to bulk read thumbnails: %s %s", err, stderr.String())
		return
	}

	thumbBytes := out.Bytes()

	offset := 0
	for _, file := range meta.Files {
		m, err := file.GetMedia()
		if err != nil {
			util.DisplayError(err)
			t.err = fmt.Errorf("failed to parse raw thumbnails, could not get media for %s", file.String())
			return
		}

		thumbLen := int(m.QueryExif(fmt.Sprintf("%sLength", meta.ExifThumbType)).(float64))
		m.DumpThumbBytes(thumbBytes[offset : offset+thumbLen])
		offset += thumbLen

	}
	util.Debug.Println("Finished loading thumbs without error for", meta.Files[0].GetParent().String())
}
