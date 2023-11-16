package dataProcess

import (
	"archive/zip"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
)

func _scan(path, username string, recursive bool) (WorkerPool) {
	scanPath := dataStore.GuaranteeUserAbsolutePath(path, username)

	_, err := os.Stat(scanPath)
	util.FailOnError(err, "Scan path does not exist")

	wp := ScanDirectory(scanPath, username, recursive)

	return wp
}

func ScanDir(meta ScanMetadata) {

	absolutePath := dataStore.GuaranteeUserAbsolutePath(meta.Path, meta.Username)
	wp := _scan(meta.Path, meta.Username, meta.Recursive)

	// TODO: ETHAN THERE IS A BUG HERE
	// THE SCAN DIR TASK DOES NOT WORK BECAUSE THE WP IT RETURNS IS EMPTY BECAUSE IT USES THE TASK TRACKER WP,
	// SO THIS EXITS INSTANTLY AND DOES NOTHING
	// MAKE IT USE ITS OWN WP AGAIN SO IT CAN BROADCAST CORRECTLY thanks <3
	var previousRemaining int
	remainingTasks, totalTasks, _, _ := wp.Status()
	for remainingTasks > 0 {
		time.Sleep(time.Second)
		remainingTasks, _, _, _ = wp.Status()

		// Don't send new message unless new data
		if remainingTasks == previousRemaining {
			continue
		} else {
			previousRemaining = remainingTasks
		}

		status := struct {RemainingTasks int `json:"remainingTasks"`; TotalTasks int `json:"totalTasks"`} {RemainingTasks: remainingTasks, TotalTasks: totalTasks}
		Broadcast("path", absolutePath, "scan_directory_progress", status)
	}
}

func ScanFile(meta ScanMetadata) {
	db := dataStore.NewDB(meta.Username)
	err := ProcessMediaFile(meta.Path, meta.Username, db)
	if err != nil {
		util.DisplayError(err, "Failed to process new meda file")
	}
}

type zipFileTuple struct {
	absPath string
	zipPath string
}

func getFilesForZip(realFile, zipPath, username string) []zipFileTuple {
	absoluteFilepath := dataStore.GuaranteeUserAbsolutePath(realFile, username)
	stat, err := os.Stat(absoluteFilepath)
	util.FailOnError(err, "Failed to get stats of file to add to zip")

	var files []zipFileTuple
	if stat.IsDir() {
		walker := func(path string, info os.FileInfo, err error) error {
			if path == absoluteFilepath {
				return nil
			}
			files = append(files, getFilesForZip(path, zipPath + realFile + "/", username)...)
			return nil
		}
		err = filepath.Walk(absoluteFilepath, walker)
		util.FailOnError(err, "")
	} else {
		zipRelativePath := zipPath + filepath.Base(absoluteFilepath)
		zipFile := zipFileTuple{absPath: absoluteFilepath, zipPath: zipRelativePath}

		return []zipFileTuple{zipFile}
	}
	return files
}

func addFileToZip(file zipFileTuple, zipWriter *zip.Writer) {
	f, err := os.Open(file.absPath)
	util.FailOnError(err, "Could not open file for adding to takeout zip")
	w, err := zipWriter.Create(file.zipPath)
	util.FailOnError(err, "")

	_, err = io.Copy(w,f)
	util.FailOnError(err, "")

	f.Close()
}

func createZipFromPaths(task *Task) {
	zipMeta := task.metadata.(ZipMetadata)
	paths := zipMeta.Paths
	username := zipMeta.Username

	takeoutHash := util.HashOfString(8, paths...)
	task.setResult(KeyVal{Key: "takeoutId", Val: takeoutHash})

	zipPath := filepath.Join(util.GetTakeoutDir(), takeoutHash + ".zip")
	_, err := os.Stat(zipPath)
	if !errors.Is(err, fs.ErrNotExist) { // If the zip file already exists, then we're done
		task.setComplete("task", "zip_complete")
		return
	}

	zippy, err := os.Create(zipPath)
	util.FailOnError(err, "Could not create zip takeout file")

	var allFilesRecursive []zipFileTuple
	for _, val := range paths {
		allFilesRecursive = append(allFilesRecursive, getFilesForZip(val, "", username)...)
	}

	defer zippy.Close()
	zipWriter := zip.NewWriter(zippy)
	defer zipWriter.Close()

	var totalFiles = len(allFilesRecursive)
	for i, file := range allFilesRecursive {
		addFileToZip(file, zipWriter)
		status := struct {RemainingFiles int `json:"remainingFiles"`; TotalFiles int `json:"totalFiles"`} {RemainingFiles: totalFiles - i, TotalFiles: totalFiles}
		Broadcast("task", task.TaskId, "create_zip_progress", status)
	}
	status := struct {RemainingFiles int `json:"remainingFiles"`; TotalFiles int `json:"totalFiles"`} {RemainingFiles: 0, TotalFiles: totalFiles}
	Broadcast("task", task.TaskId, "create_zip_progress", status)
	task.setComplete("task", "zip_complete")
}