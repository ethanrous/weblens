package proxy

import (
	"fmt"
	"io"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
)

var _ models.FileService = (*ProxyFileService)(nil)

type ProxyFileService struct {
	Core *models.Instance
}

func (pfs *ProxyFileService) GetZip(id fileTree.FileId) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) GetMediaRoot() *fileTree.WeblensFileImpl {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) PathToFile(
	searchPath string,
) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) CreateFile(parent *fileTree.WeblensFileImpl, filename string) (
	*fileTree.WeblensFileImpl, error,
) {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) CreateFolder(
	parent *fileTree.WeblensFileImpl, foldername string, caster models.FileCaster,
) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) GetFile(id fileTree.FileId) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) GetFiles(ids []fileTree.FileId) ([]*fileTree.WeblensFileImpl, error) {
	if len(ids) == 0 {
		return []*fileTree.WeblensFileImpl{}, nil
	}
	return CallHomeStruct[[]*fileTree.WeblensFileImpl](pfs.Core, "POST", "/files", ids)
}

func (pfs *ProxyFileService) GetFileSafe(
	id fileTree.FileId, accessor *models.User, share *models.FileShare,
) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) GetFileOwner(file *fileTree.WeblensFileImpl) *models.User {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) IsFileInTrash(file *fileTree.WeblensFileImpl) bool {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) ImportFile(f *fileTree.WeblensFileImpl) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) RenameFile(
	file *fileTree.WeblensFileImpl, newName string, caster models.FileCaster,
) error {
	panic("implement me")
}

func (pfs *ProxyFileService) MoveFiles(
	files []*fileTree.WeblensFileImpl, destFolder *fileTree.WeblensFileImpl, caster models.FileCaster,
) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) MoveFilesToTrash(
	file []*fileTree.WeblensFileImpl, mover *models.User, share *models.FileShare, caster models.FileCaster,
) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) ReturnFilesFromTrash(files []*fileTree.WeblensFileImpl, caster models.FileCaster) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) PermanentlyDeleteFiles(files []*fileTree.WeblensFileImpl, caster models.FileCaster) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) ReadFile(f *fileTree.WeblensFileImpl) (io.ReadCloser, error) {
	resp, err := callHome(pfs.Core, "GET", fmt.Sprintf("/file/%s/content", f.ID()), nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (pfs *ProxyFileService) GetThumbFileName(filename string) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) NewCacheFile(
	contentId string, quality models.MediaQuality, pageNum int,
) (fileTree.WeblensFile, error) {
	return nil, nil
}

func (pfs *ProxyFileService) DeleteCacheFile(file fileTree.WeblensFile) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) AddTask(f *fileTree.WeblensFileImpl, t *task.Task) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) RemoveTask(f *fileTree.WeblensFileImpl, t *task.Task) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) GetTasks(f *fileTree.WeblensFileImpl) []*task.Task {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) GetMediaJournal() fileTree.Journal {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) ResizeDown(file *fileTree.WeblensFileImpl, caster models.FileCaster) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) ResizeUp(file *fileTree.WeblensFileImpl, caster models.FileCaster) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) NewZip(zipName string, owner *models.User) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}
