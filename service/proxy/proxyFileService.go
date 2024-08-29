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

func (pfs *ProxyFileService) GetFile(id fileTree.FileId) (*fileTree.WeblensFile, error) {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) GetFiles(ids []fileTree.FileId) ([]*fileTree.WeblensFile, error) {
	if len(ids) == 0 {
		return []*fileTree.WeblensFile{}, nil
	}
	return callHomeStruct[[]*fileTree.WeblensFile](pfs.Core, "POST", "/files", ids)
}

func (pfs *ProxyFileService) GetFileSafe(
	id fileTree.FileId, accessor *models.User, share *models.FileShare,
) (*fileTree.WeblensFile, error) {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) GetFileOwner(file *fileTree.WeblensFile) *models.User {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) IsFileInTrash(file *fileTree.WeblensFile) bool {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) ImportFile(f *fileTree.WeblensFile) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) RenameFile(file *fileTree.WeblensFile, newName string, caster models.FileCaster) error {
	panic("implement me")
}

func (pfs *ProxyFileService) MoveFiles(
	files []*fileTree.WeblensFile, destFolder *fileTree.WeblensFile, caster models.FileCaster,
) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) MoveFileToTrash(
	file *fileTree.WeblensFile, mover *models.User, share *models.FileShare, caster models.FileCaster,
) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) ReturnFilesFromTrash(files []*fileTree.WeblensFile, caster models.FileCaster) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) PermanentlyDeleteFiles(files []*fileTree.WeblensFile, caster models.FileCaster) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) ReadFile(f *fileTree.WeblensFile) (io.ReadCloser, error) {
	resp, err := callHome(pfs.Core, "GET", fmt.Sprintf("/file/%s/content", f.ID()), nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (pfs *ProxyFileService) GetThumbFileName(filename string) (*fileTree.WeblensFile, error) {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) NewCacheFile(
	contentId string, quality models.MediaQuality, pageNum int,
) (*fileTree.WeblensFile, error) {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) DeleteCacheFile(file *fileTree.WeblensFile) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) AddTask(f *fileTree.WeblensFile, t *task.Task) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) RemoveTask(f *fileTree.WeblensFile, t *task.Task) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) GetTasks(f *fileTree.WeblensFile) []*task.Task {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) GetMediaJournal() fileTree.JournalService {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) ResizeDown(file *fileTree.WeblensFile, caster models.FileCaster) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) ResizeUp(file *fileTree.WeblensFile, caster models.FileCaster) error {
	// TODO implement me
	panic("implement me")
}

func (pfs *ProxyFileService) NewZip(zipName string, owner *models.User) (*fileTree.WeblensFile, error) {
	// TODO implement me
	panic("implement me")
}
