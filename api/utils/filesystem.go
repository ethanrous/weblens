package util

import (
	"path/filepath"
	"strings"
)

func AbsoluteToRelativePath(absolutePath string) (string) {
	absolutePrefix := GetMediaRoot()
	relativePath := strings.Replace(absolutePath, absolutePrefix, "", -1)
	return relativePath
}

func RelativeToAbsolutePath(relativePath string) (string) {
	absolutePrefix := GetMediaRoot()
	absolutePath := filepath.Join(absolutePrefix, relativePath)
	return absolutePath
}