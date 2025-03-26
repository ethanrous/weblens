package file

import (
	"context"
	"os"
)

const tmpPattern = "weblens-*.tmp"

func NewTmpFile() (*os.File, error) {
	return os.CreateTemp("", tmpPattern)
}

func Hardlink(ctx context.Context, sourcePath, destPath string) error {
	return os.Link(sourcePath, destPath)
}

func DeleteFile(ctx context.Context, filePath string) error {
	return os.Remove(filePath)
}
