package mock

import (
	"io"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
)

var _ models.FileService = (*MockFileService)(nil)

type MockFileService struct{}

func (mfs *MockFileService) GetZip(id fileTree.FileId) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

var usernames = []string{
	"Rosalie Haas",
	"Morris Coleman",
	"Cathy Shelton",
	"Gustavo Gould",
	"Israel Lee",
	"Bob Mayo",
	"Lora Massey",
	"Pam Silva",
}

func (mfs *MockFileService) GetMediaRoot() *fileTree.WeblensFileImpl {
	// TODO implement me
	panic("implement me")
}

func (mfs *MockFileService) PathToFile(searchPath string) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (mfs *MockFileService) CreateFile(parent *fileTree.WeblensFileImpl, filename string) (
	*fileTree.WeblensFileImpl, error,
) {
	// TODO implement me
	panic("implement me")
}

func (mfs *MockFileService) CreateFolder(
	parent *fileTree.WeblensFileImpl, foldername string, caster models.FileCaster,
) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (mfs *MockFileService) GetFile(id fileTree.FileId) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (mfs *MockFileService) GetFiles(ids []fileTree.FileId) ([]*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (mfs *MockFileService) GetFileSafe(
	id fileTree.FileId, accessor *models.User, share *models.FileShare,
) (*fileTree.WeblensFileImpl, error) {
	return nil, nil
}

func (mfs *MockFileService) GetFileOwner(file *fileTree.WeblensFileImpl) *models.User {
	return nil
}

func (mfs *MockFileService) IsFileInTrash(file *fileTree.WeblensFileImpl) bool {
	return false
}

func (mfs *MockFileService) ImportFile(f *fileTree.WeblensFileImpl) error {
	return nil
}

func (mfs *MockFileService) MoveFiles(
	files []*fileTree.WeblensFileImpl, destFolder *fileTree.WeblensFileImpl, caster models.FileCaster,
) error {
	return nil
}

func (mfs *MockFileService) RenameFile(file *fileTree.WeblensFileImpl, newName string, caster models.FileCaster) error {
	panic("implement me")
}

func (mfs *MockFileService) MoveFilesToTrash(
	file []*fileTree.WeblensFileImpl, mover *models.User, share *models.FileShare, caster models.FileCaster,
) error {
	return nil
}

func (mfs *MockFileService) ReturnFilesFromTrash(files []*fileTree.WeblensFileImpl, caster models.FileCaster) error {
	return nil
}

func (mfs *MockFileService) PermanentlyDeleteFiles(files []*fileTree.WeblensFileImpl, caster models.FileCaster) error {
	return nil
}

func (mfs *MockFileService) ReadFile(file *fileTree.WeblensFileImpl) (io.ReadCloser, error) {
	return nil, nil
}

func (mfs *MockFileService) AddTask(f *fileTree.WeblensFileImpl, t *task.Task) error {
	return nil
}

func (mfs *MockFileService) RemoveTask(f *fileTree.WeblensFileImpl, t *task.Task) error {
	return nil
}

func (mfs *MockFileService) GetTasks(f *fileTree.WeblensFileImpl) []*task.Task {
	return nil
}

func (mfs *MockFileService) GetMediaJournal() fileTree.Journal {
	return nil
}

func (mfs *MockFileService) ResizeDown(file *fileTree.WeblensFileImpl, caster models.FileCaster) error {
	return nil
}

func (mfs *MockFileService) ResizeUp(file *fileTree.WeblensFileImpl, caster models.FileCaster) error {
	return nil
}

func (mfs *MockFileService) NewZip(zipName string, owner *models.User) (*fileTree.WeblensFileImpl, error) {
	return nil, nil
}

func (mfs *MockFileService) DeleteCacheFile(file fileTree.WeblensFile) error {
	return nil
}

func (mfs *MockFileService) GetMediaCacheByFilename(filename string) (*fileTree.WeblensFileImpl, error) {
	return nil, nil
}

func (mfs *MockFileService) NewCacheFile(
	media *models.Media, quality models.MediaQuality, pageNum int,
) (*fileTree.WeblensFileImpl, error) {
	filename := media.FmtCacheFileName(quality, pageNum)

	cache := fileTree.NewWeblensFile("TODO", filename, nil, false)
	cache.SetMemOnly(true)
	return cache, nil
}
