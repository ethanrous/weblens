package util

import (
	"path/filepath"
	"strings"
)

func GuaranteeRelativePath(absolutePath string) (string) {
	absolutePrefix := GetMediaRoot()
	relativePath := strings.TrimPrefix(absolutePath, absolutePrefix)
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