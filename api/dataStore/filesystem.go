package dataStore

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethrousseau/weblens/api/util"
)

func GuaranteeRelativePath(absolutePath string) (string) {
	absolutePrefix := util.GetMediaRoot()
	relativePath := filepath.Join("/", strings.TrimPrefix(absolutePath, absolutePrefix))
	return relativePath
}

func GuaranteeAbsolutePath(relativePath string) (string) {
	if isAbsolutePath(relativePath) {
		return relativePath
	}
	absolutePrefix := util.GetMediaRoot()
	absolutePath := filepath.Join(absolutePrefix, relativePath)
	return absolutePath
}

func isAbsolutePath(mysteryPath string) (bool) {
	return strings.HasPrefix(mysteryPath, util.GetMediaRoot())
}

func AddFileToZip(realFile, zipPath string, zipWriter *zip.Writer) {
	absoluteFilepath := GuaranteeAbsolutePath(realFile)
	util.Debug.Printf("Adding %s to zip file", absoluteFilepath)

	stat, err := os.Stat(absoluteFilepath)
	util.FailOnError(err, "Failed to get stats of file to add to zip")

	if stat.IsDir() {
		walker := func(path string, info os.FileInfo, err error) error {
			util.Debug.Println("Walking found ", path)
			if path == absoluteFilepath {
				return nil
			}
			AddFileToZip(path, zipPath + realFile + "/", zipWriter)
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

func CreateZipFromPaths(paths []string) string {
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
		AddFileToZip(val, "", zipWriter)
	}

	return takeoutHash
}