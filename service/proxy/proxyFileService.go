package proxy

import (
	"fmt"
	"io"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
)

var _ models.FileService = (*ProxyFileService)(nil)

type ProxyFileService struct {
	Core *models.Instance
}

func (pfs *ProxyFileService) GetZip(id fileTree.FileId) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (pfs *ProxyFileService) GetMediaRoot() *fileTree.WeblensFileImpl {

	panic("implement me")
}

func (pfs *ProxyFileService) PathToFile(
	searchPath string,
) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (pfs *ProxyFileService) CreateFile(parent *fileTree.WeblensFileImpl, filename string, event *fileTree.FileEvent) (
	*fileTree.WeblensFileImpl, error,
) {

	panic("implement me")
}

func (pfs *ProxyFileService) CreateTmpFile(parent *fileTree.WeblensFileImpl, filename string) (
	*fileTree.WeblensFileImpl, error,
) {

	panic("implement me")
}

func (pfs *ProxyFileService) CreateFolder(
	parent *fileTree.WeblensFileImpl, foldername string, caster models.FileCaster,
) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (pfs *ProxyFileService) GetUserFile(id fileTree.FileId) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (pfs *ProxyFileService) GetFiles(ids []fileTree.FileId) ([]*fileTree.WeblensFileImpl, error) {
	if len(ids) == 0 {
		return []*fileTree.WeblensFileImpl{}, nil
	}

	r := NewRequest(pfs.Core, "POST", "/files").WithBody(ids)
	return CallHomeStruct[[]*fileTree.WeblensFileImpl](r)
}

func (pfs *ProxyFileService) GetFileSafe(
	id fileTree.FileId, accessor *models.User, share *models.FileShare,
) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (pfs *ProxyFileService) GetFileOwner(file *fileTree.WeblensFileImpl) *models.User {

	panic("implement me")
}

func (pfs *ProxyFileService) IsFileInTrash(file *fileTree.WeblensFileImpl) bool {

	panic("implement me")
}

func (pfs *ProxyFileService) ImportFile(f *fileTree.WeblensFileImpl) error {

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

	panic("implement me")
}

func (pfs *ProxyFileService) MoveFilesToTrash(
	file []*fileTree.WeblensFileImpl, mover *models.User, share *models.FileShare, caster models.FileCaster,
) error {

	panic("implement me")
}

func (pfs *ProxyFileService) ReturnFilesFromTrash(files []*fileTree.WeblensFileImpl, caster models.FileCaster) error {

	panic("implement me")
}

func (pfs *ProxyFileService) DeleteFiles(files []*fileTree.WeblensFileImpl, caster models.FileCaster) error {

	panic("implement me")
}

func (pfs *ProxyFileService) RestoreFiles(
	ids []fileTree.FileId, newParent *fileTree.WeblensFileImpl, restoreTime time.Time, caster models.FileCaster,
) error {

	panic("implement me")
}

func (pfs *ProxyFileService) ReadFile(f *fileTree.WeblensFileImpl) (io.ReadCloser, error) {
	resp, err := NewRequest(pfs.Core, "GET", fmt.Sprintf("/file/%s/content", f.ID())).Call()
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (pfs *ProxyFileService) GetMediaCacheByFilename(filename string) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (pfs *ProxyFileService) NewCacheFile(
	media *models.Media, quality models.MediaQuality, pageNum int,
) (*fileTree.WeblensFileImpl, error) {
	return nil, nil
}

func (pfs *ProxyFileService) DeleteCacheFile(file fileTree.WeblensFile) error {

	panic("implement me")
}

func (pfs *ProxyFileService) AddTask(f *fileTree.WeblensFileImpl, t *task.Task) error {

	panic("implement me")
}

func (pfs *ProxyFileService) RemoveTask(f *fileTree.WeblensFileImpl, t *task.Task) error {

	panic("implement me")
}

func (pfs *ProxyFileService) GetTasks(f *fileTree.WeblensFileImpl) []*task.Task {

	panic("implement me")
}

func (pfs *ProxyFileService) GetUsersJournal() fileTree.Journal {

	panic("implement me")
}

func (pfs *ProxyFileService) ResizeDown(file *fileTree.WeblensFileImpl, caster models.FileCaster) error {

	panic("implement me")
}

func (pfs *ProxyFileService) ResizeUp(file *fileTree.WeblensFileImpl, caster models.FileCaster) error {

	panic("implement me")
}

func (pfs *ProxyFileService) NewZip(zipName string, owner *models.User) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}
