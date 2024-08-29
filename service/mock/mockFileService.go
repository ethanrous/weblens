package mock

import (
	"io"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
)

var _ models.FileService = (*MockFileService)(nil)

type MockFileService struct{}

func (mfs *MockFileService) GetFile(id fileTree.FileId) (*fileTree.WeblensFile, error) {
	// TODO implement me
	panic("implement me")
}

func (mfs *MockFileService) GetFiles(ids []fileTree.FileId) ([]*fileTree.WeblensFile, error) {
	// TODO implement me
	panic("implement me")
}

func (mfs *MockFileService) GetFileSafe(
	id fileTree.FileId, accessor *models.User, share *models.FileShare,
) (*fileTree.WeblensFile, error) {
	return nil, nil
}

func (mfs *MockFileService) GetFileOwner(file *fileTree.WeblensFile) *models.User {
	return nil
}

func (mfs *MockFileService) IsFileInTrash(file *fileTree.WeblensFile) bool {
	return false
}

func (mfs *MockFileService) ImportFile(f *fileTree.WeblensFile) error {
	return nil
}

func (mfs *MockFileService) MoveFiles(
	files []*fileTree.WeblensFile, destFolder *fileTree.WeblensFile, caster models.FileCaster,
) error {
	return nil
}

func (mfs *MockFileService) RenameFile(file *fileTree.WeblensFile, newName string, caster models.FileCaster) error {
	panic("implement me")
}

func (mfs *MockFileService) MoveFileToTrash(
	file *fileTree.WeblensFile, mover *models.User, share *models.FileShare, caster models.FileCaster,
) error {
	return nil
}

func (mfs *MockFileService) ReturnFilesFromTrash(files []*fileTree.WeblensFile, caster models.FileCaster) error {
	return nil
}

func (mfs *MockFileService) PermanentlyDeleteFiles(files []*fileTree.WeblensFile, caster models.FileCaster) error {
	return nil
}

func (mfs *MockFileService) ReadFile(file *fileTree.WeblensFile) (io.ReadCloser, error) {
	return nil, nil
}

func (mfs *MockFileService) AddTask(f *fileTree.WeblensFile, t *task.Task) error {
	return nil
}

func (mfs *MockFileService) RemoveTask(f *fileTree.WeblensFile, t *task.Task) error {
	return nil
}

func (mfs *MockFileService) GetTasks(f *fileTree.WeblensFile) []*task.Task {
	return nil
}

func (mfs *MockFileService) GetMediaJournal() fileTree.JournalService {
	return nil
}

func (mfs *MockFileService) ResizeDown(file *fileTree.WeblensFile, caster models.FileCaster) error {
	return nil
}

func (mfs *MockFileService) ResizeUp(file *fileTree.WeblensFile, caster models.FileCaster) error {
	return nil
}

func (mfs *MockFileService) NewZip(zipName string, owner *models.User) (*fileTree.WeblensFile, error) {
	return nil, nil
}

func (mfs *MockFileService) DeleteCacheFile(file *fileTree.WeblensFile) error {
	return nil
}

func (mfs *MockFileService) GetThumbFileName(filename string) (*fileTree.WeblensFile, error) {
	return nil, nil
}

func (mfs *MockFileService) NewCacheFile(contentId string, quality models.MediaQuality, pageNum int) (
	*fileTree.WeblensFile,
	error,
) {
	return nil, nil
}
