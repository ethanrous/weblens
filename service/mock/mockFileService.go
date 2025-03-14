package mock

import (
	"io"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/task"
)

var _ models.FileService = (*MockFileService)(nil)

type MockFileService struct {
	trees map[string]fileTree.FileTree
}

func NewMockFileService() *MockFileService {
	return &MockFileService{
		trees: make(map[string]fileTree.FileTree),
	}
}

func (mfs *MockFileService) GetZip(id fileTree.FileId) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (mfs *MockFileService) Size(treeName string) int64 {
	panic("implement me")
}

func (mfs *MockFileService) AddTree(tree fileTree.FileTree) {
	mfs.trees[tree.GetRoot().GetPortablePath().RootName()] = tree
}

func (mfs *MockFileService) GetUsersRoot() *fileTree.WeblensFileImpl {

	panic("implement me")
}

func (mfs *MockFileService) PathToFile(searchPath string) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (mfs *MockFileService) CreateFile(parent *fileTree.WeblensFileImpl, filename string, event *fileTree.FileEvent, caster models.FileCaster) (
	*fileTree.WeblensFileImpl, error,
) {

	panic("implement me")
}

func (mfs *MockFileService) CreateFolder(
	parent *fileTree.WeblensFileImpl, foldername string, event *fileTree.FileEvent, caster models.FileCaster,
) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (mfs *MockFileService) CreateUserHome(user *models.User) error {
	home := fileTree.NewWeblensFile(user.Username, user.Username, nil, true)
	user.SetHomeFolder(home)
	trash := fileTree.NewWeblensFile(user.Username+"trash", ".user_trash", home, true)
	user.SetTrashFolder(trash)

	return nil
}

func (mfs *MockFileService) CreateRestoreFile(lifetime *fileTree.Lifetime) (
	restoreFile *fileTree.WeblensFileImpl, err error,
) {
	panic("implement me")
}

func (mfs *MockFileService) GetFileByTree(id fileTree.FileId, treeAlias string) (*fileTree.WeblensFileImpl, error) {
	return mfs.trees[treeAlias].Get(id), nil
}

func (mfs *MockFileService) GetFileByContentId(contentId models.ContentId) (*fileTree.WeblensFileImpl, error) {
	panic("implement me")
}

func (mfs *MockFileService) GetFiles(ids []fileTree.FileId) ([]*fileTree.WeblensFileImpl, []fileTree.FileId, error) {

	panic("implement me")
}

func (mfs *MockFileService) GetFileSafe(
	id fileTree.FileId, accessor *models.User, share *models.FileShare,
) (*fileTree.WeblensFileImpl, error) {
	return nil, nil
}

func (mfs *MockFileService) GetFileTreeByName(treeName string) fileTree.FileTree {
	return nil
}

func (mfs *MockFileService) GetFileOwner(file *fileTree.WeblensFileImpl) (*models.User, error) {
	return &models.User{
		Username: "MOCK_USER",
	}, nil
}

func (mfs *MockFileService) IsFileInTrash(file *fileTree.WeblensFileImpl) bool {
	return false
}

func (mfs *MockFileService) ImportFile(f *fileTree.WeblensFileImpl) error {
	return nil
}

func (mfs *MockFileService) MoveFiles(
	files []*fileTree.WeblensFileImpl, destFolder *fileTree.WeblensFileImpl, treeName string, caster models.FileCaster,
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

func (mfs *MockFileService) DeleteFiles(files []*fileTree.WeblensFileImpl, treeName string, caster models.FileCaster) error {
	return nil
}

func (mfs *MockFileService) RestoreFiles(
	ids []fileTree.FileId, newParent *fileTree.WeblensFileImpl, restoreTime time.Time, caster models.FileCaster,
) error {

	panic("implement me")
}

func (mfs *MockFileService) RestoreHistory(lifetimes []*fileTree.Lifetime) error {

	panic("implement me")
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

func (mfs *MockFileService) GetJournalByTree(treeName string) fileTree.Journal {
	return mfs.trees[treeName].GetJournal()
}

func (mfs *MockFileService) ResizeDown(file *fileTree.WeblensFileImpl, event *fileTree.FileEvent, caster models.FileCaster) error {
	return nil
}

func (mfs *MockFileService) ResizeUp(file *fileTree.WeblensFileImpl, event *fileTree.FileEvent, caster models.FileCaster) error {
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

func (mfs *MockFileService) GetThumbsDir() (*fileTree.WeblensFileImpl, error) {
	return nil, nil
}

func (mfs *MockFileService) NewBackupFile(lt *fileTree.Lifetime) (*fileTree.WeblensFileImpl, error) {
	return nil, nil
}

func (mfs *MockFileService) GetFolderCover(folder *fileTree.WeblensFileImpl) (models.ContentId, error) {
	return "", nil
}

func (mfs *MockFileService) SetFolderCover(folderId fileTree.FileId, coverId models.ContentId) error {
	return nil
}

func (mfs *MockFileService) UserPathToFile(searchPath string, user *models.User) (*fileTree.WeblensFileImpl, error) {
	return nil, nil
}
