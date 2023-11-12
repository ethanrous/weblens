package dataStore

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

type FileInfo struct{
	Imported bool `json:"imported"` // If the item has been loaded into the database, dictates if MediaData is set or not
	IsDir bool `json:"isDir"`
	Size int `json:"size"`
	ModTime time.Time `json:"modTime"`
	Filepath string `json:"filepath"`
	MediaData Media `json:"mediaData"`
}


// Take a (possibly) absolutePath (string), and return a path to the same location, relative to media root (from .env)
func GuaranteeRelativePath(absolutePath string) (string) {
	absolutePrefix := util.GetMediaRoot()
	relativePath := filepath.Join("/", strings.TrimPrefix(absolutePath, absolutePrefix))
	return relativePath
}

// Take a possibly absolute `path` (string), and return a path to the same location, relative to the given users home directory
// Returns an error if the file is not in the users home directory, or tries to access the "SYS" home directory, which does not exist
func GuaranteeUserRelativePath(path, username string) (string, error) {
	if username == "SYS" {
		return "", fmt.Errorf("attempt to get relative path with SYS user")
	}

	absolutePrefix := filepath.Join(util.GetMediaRoot(), username)
	if isAbsolutePath(path) && !strings.HasPrefix(path, absolutePrefix) {
		util.Debug.Println("Prefix:", absolutePrefix)
		return "", fmt.Errorf("attempt to get user relative path for a file not in user's home directory\n File: %s\nUser: %s", path, username)
	}

	relativePath := filepath.Join("/", strings.TrimPrefix(path, absolutePrefix))
	return relativePath, nil
}

func GuaranteeAbsolutePath(relativePath, username string) (string) {
	if username == "SYS" {
		util.Error.Panicln("Attempt to get absolute path with SYS user")
	}

	if isAbsolutePath(relativePath) {
		return relativePath
	}
	// absolutePrefix := filepath.Join(util.GetMediaRoot(), username)
	absolutePrefix := util.GetMediaRoot()
	absolutePath := filepath.Join(absolutePrefix, relativePath)
	return absolutePath
}

func GuaranteeUserAbsolutePath(relativePath, username string) (string) {
	if username == "SYS" {
		util.Error.Panicln("Attempt to get absolute path with SYS user")
	}

	if isAbsolutePath(relativePath) {
		return relativePath
	}

	relativePath = strings.TrimPrefix(relativePath, "/" + username)

	absolutePrefix := filepath.Join(util.GetMediaRoot(), username)
	absolutePath := filepath.Join(absolutePrefix, relativePath)
	return absolutePath
}

func isAbsolutePath(mysteryPath string) (bool) {
	return strings.HasPrefix(mysteryPath, util.GetMediaRoot())
}

func GetOwnerFromFilepath(path string) string {
	relativePath := GuaranteeRelativePath(path)
	var username string
	if strings.Index(relativePath, "/") == 0 {
		username = relativePath[1:strings.Index(relativePath[1:], "/") + 1]
	} else {
		username = relativePath[:strings.Index(relativePath, "/") + 1]
	}

	return username
}

func AddFileToZip(realFile, zipPath, username string, zipWriter *zip.Writer) {
	absoluteFilepath := GuaranteeAbsolutePath(realFile, username)
	util.Debug.Printf("Adding %s to zip file", absoluteFilepath)

	stat, err := os.Stat(absoluteFilepath)
	util.FailOnError(err, "Failed to get stats of file to add to zip")

	if stat.IsDir() {
		walker := func(path string, info os.FileInfo, err error) error {
			util.Debug.Println("Walking found ", path)
			if path == absoluteFilepath {
				return nil
			}
			AddFileToZip(path, zipPath + realFile + "/", username, zipWriter)
			return nil
		}
		err = filepath.Walk(absoluteFilepath, walker)
		util.FailOnError(err, "")
	} else {
		f, err := os.Open(absoluteFilepath)
		util.FailOnError(err, "Could not open file for adding to takeout zip")

		zipRelativePath := zipPath + filepath.Base(absoluteFilepath)

		w, err := zipWriter.Create(zipRelativePath)
		util.FailOnError(err, "")

		_, err = io.Copy(w,f)
		util.FailOnError(err, "")

		f.Close()
	}

}

func CreateZipFromPaths(paths []string, username string) string {
	var preString string
	for _, val := range paths {
		preString += val
	}
	takeoutHash := util.HashOfString(preString, 8)

	zipPath := fmt.Sprintf("%s/%s.zip", util.GetTakeoutDir(), takeoutHash)
	_, err := os.Stat(zipPath)
	if !errors.Is(err, os.ErrNotExist) {
		return takeoutHash
	}

	zippy, err := os.Create(zipPath)
	util.FailOnError(err, "Could not create zip takeout file")
	defer zippy.Close()

	zipWriter := zip.NewWriter(zippy)
	defer zipWriter.Close()

	for _, val := range paths {
		AddFileToZip(val, "", username, zipWriter)
	}

	return takeoutHash
}

var dirIgnore = []string{
	".DS_Store",
}

func FormatFileInfo(p, username string, db Weblensdb) (FileInfo, bool) {

	// absolutePath := dataStore.GuaranteeAbsolutePath(parentDir, username)
	absolutePath := GuaranteeUserAbsolutePath(p, username)
	relativePath, _ := GuaranteeUserRelativePath(absolutePath, username)

	file, err := os.Stat(p)
	util.FailOnError(err, "Failed to format file info")

	var formattedInfo FileInfo
	var include bool = false

	if !slices.Contains(dirIgnore, file.Name()) {
		include = true

		var filled bool
		var fileSize int64
		var mediaData Media

		if !file.IsDir() {
			mediaData, _ = db.GetMediaByFilepath(absolutePath, true)
			// if err != nil {
				// meta := dataProcess.ScanMetadata{Path: absolutePath, Username: username, Recursive: false}
				// metaS, _ := json.Marshal(meta)

				// task := dataProcess.Task{TaskType: "scan_directory", Metadata: string(metaS)}
				// dataProcess.RequestTask(task)

			// }
			// util.FailOnError(err, "Failed to format file info")

			filled, _ = mediaData.IsFilledOut(false)
			mediaData.Thumbnail64 = ""

			fileSize = file.Size()

		} else {
			fileSize, err = util.DirSize(absolutePath)
			util.FailOnError(err, "Failed to get dir size")
		}

		formattedInfo = FileInfo{Imported: filled, IsDir: file.IsDir(), Size: int(fileSize), ModTime: file.ModTime(), Filepath: relativePath, MediaData: mediaData}
	}
	return formattedInfo, include
}