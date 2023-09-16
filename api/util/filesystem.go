package util

import (
	"path/filepath"
	"strings"
)

func GuaranteeRelativePath(absolutePath string) (string) {
	absolutePrefix := getMediaRoot()
	relativePath := filepath.Join("/", strings.TrimPrefix(absolutePath, absolutePrefix))
	return relativePath
}

func GuaranteeAbsolutePath(relativePath string) (string) {
	if isAbsolutePath(relativePath) {
		return relativePath
	}
	absolutePrefix := getMediaRoot()
	absolutePath := filepath.Join(absolutePrefix, relativePath)
	return absolutePath
}

func isAbsolutePath(mysteryPath string) (bool) {
	return strings.HasPrefix(mysteryPath, getMediaRoot())
}