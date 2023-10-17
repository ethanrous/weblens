package util

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func GuaranteeRelativePath(absolutePath string) (string) {
	absolutePrefix := GetMediaRoot()
	relativePath := filepath.Join("/", strings.TrimPrefix(absolutePath, absolutePrefix))
	return relativePath
}

func GuaranteeAbsolutePath(relativePath string) (string) {
	if isAbsolutePath(relativePath) {
		return relativePath
	}
	absolutePrefix := GetMediaRoot()
	absolutePath := filepath.Join(absolutePrefix, relativePath)
	return absolutePath
}

func isAbsolutePath(mysteryPath string) (bool) {
	return strings.HasPrefix(mysteryPath, GetMediaRoot())
}

func AddFileToZip(realFile, zipPath string, zipWriter *zip.Writer) {
	absoluteFilepath := GuaranteeAbsolutePath(realFile)
	Debug.Printf("Adding %s to zip file", absoluteFilepath)

	stat, err := os.Stat(absoluteFilepath)
	FailOnError(err, "Failed to get stats of file to add to zip")

	if stat.IsDir() {
		walker := func(path string, info os.FileInfo, err error) error {
			Debug.Println("Walking found ", path)
			if path == absoluteFilepath {
				return nil
			}
			AddFileToZip(path, zipPath + realFile + "/", zipWriter)
			return nil
		}
		err = filepath.Walk(absoluteFilepath, walker)
		FailOnError(err, "")
	} else {
		f, err := os.Open(absoluteFilepath)
		FailOnError(err, "Could not open file for adding to takeout zip")

		zipRelativePath := zipPath + filepath.Base(absoluteFilepath)

		w, err := zipWriter.Create(zipRelativePath)
		FailOnError(err, "")

		_, err = io.Copy(w,f)
		FailOnError(err, "")

		f.Close()
	}

}

func CreateZipFromPaths(paths []string) string {
	var preString string
	for _, val := range paths {
		preString += val
	}
	takeoutHash := HashOfString(preString, 8)

	zipPath := fmt.Sprintf("%s/%s.zip", GetTakeoutDir(), takeoutHash)
	_, err := os.Stat(zipPath)
	if !errors.Is(err, os.ErrNotExist) {
		return takeoutHash
	}

	zippy, err := os.Create(zipPath)
	FailOnError(err, "Could not create zip takeout file")
	defer zippy.Close()

	zipWriter := zip.NewWriter(zippy)
	defer zipWriter.Close()

	for _, val := range paths {
		AddFileToZip(val, "", zipWriter)
	}

	return takeoutHash
}