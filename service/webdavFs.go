package service

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/models"
	"golang.org/x/net/webdav"
)

var _ webdav.FileSystem = (*WebdavFs)(nil)

type WebdavFs struct {
	WeblensFs models.FileService
	Caster    models.FileCaster
}

func (w WebdavFs) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	unescapeName, err := url.QueryUnescape(name)
	if err != nil {
		return err
	}

	parent, err := w.WeblensFs.PathToFile(filepath.Dir(unescapeName))
	if err != nil {
		return err
	}

	_, err = w.WeblensFs.CreateFile(parent, filepath.Base(unescapeName))
	if err != nil {
		return err
	}

	return nil
}

func (w WebdavFs) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	unescapeName, err := url.QueryUnescape(name)
	if err != nil {
		return nil, err
	}

	// fileName := filepath.Base(unescapeName)
	// if strings.HasPrefix(fileName, "._") {
	// 	fileName = fileName[2:]
	// 	unescapeName = filepath.Dir(unescapeName) + "/" + fileName
	// }
	// if unescapeName == "." {
	// 	unescapeName = "/"
	// }

	f, err := w.WeblensFs.PathToFile(unescapeName)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (w WebdavFs) RemoveAll(ctx context.Context, name string) error {
	// TODO implement me
	panic("implement me")
}

func (w WebdavFs) Rename(ctx context.Context, oldName, newName string) error {
	unescapeOldName, err := url.QueryUnescape(oldName)
	if err != nil {
		return err
	}
	unescapeNewName, err := url.QueryUnescape(newName)
	if err != nil {
		return err
	}

	oldFile, err := w.WeblensFs.PathToFile(unescapeOldName)
	if err != nil {
		return err
	}

	newParent, err := w.WeblensFs.PathToFile(filepath.Dir(unescapeNewName))
	if err != nil {
		return err
	}

	err = w.WeblensFs.MoveFiles([]*fileTree.WeblensFileImpl{oldFile}, newParent, w.Caster)
	if err != nil {
		return err
	}

	return nil
}

func (w WebdavFs) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	unescapeName, err := url.QueryUnescape(name)
	if err != nil {
		return nil, err
	}

	fileName := filepath.Base(unescapeName)
	if strings.HasPrefix(fileName, "._") {
		fileName = fileName[2:]
		unescapeName = filepath.Dir(unescapeName) + "/" + fileName
	}

	f, err := w.WeblensFs.PathToFile(unescapeName)
	if err != nil {
		return nil, err
	}

	return f, nil
}
