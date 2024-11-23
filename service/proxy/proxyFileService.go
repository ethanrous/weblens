package proxy

import (
	"fmt"
	"io"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/task"
)

var _ models.FileService = (*ProxyFileService)(nil)

type ProxyFileService struct {
	Core *models.Instance
}

func (pfs *ProxyFileService) Size(treeName string) int64 {
	panic("implement me")
}

func (pfs *ProxyFileService) AddTree(tree fileTree.FileTree) {
	panic("implement me")
}

func (pfs *ProxyFileService) GetZip(id fileTree.FileId) (*fileTree.WeblensFileImpl, error) {
	panic("implement me")
}

func (pfs *ProxyFileService) GetUsersRoot() *fileTree.WeblensFileImpl {

	panic("implement me")
}

func (pfs *ProxyFileService) PathToFile(
	searchPath string,
) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (pfs *ProxyFileService) CreateFile(parent *fileTree.WeblensFileImpl, filename string, event *fileTree.FileEvent, caster models.FileCaster) (
	*fileTree.WeblensFileImpl, error,
) {

	panic("implement me")
}

func (pfs *ProxyFileService) CreateFolder(
	parent *fileTree.WeblensFileImpl, foldername string, event *fileTree.FileEvent, caster models.FileCaster,
) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (pfs *ProxyFileService) CreateTmpFile(parent *fileTree.WeblensFileImpl, filename string) (
	*fileTree.WeblensFileImpl, error,
) {

	panic("implement me")
}

func (pfs *ProxyFileService) CreateUserHome(user *models.User) error {
	panic("implement me")
}

func (pfs *ProxyFileService) CreateRestoreFile(lifetime *fileTree.Lifetime) (
	restoreFile *fileTree.WeblensFileImpl, err error,
) {
	panic("implement me")
}

func (pfs *ProxyFileService) GetFileByTree(id fileTree.FileId, treeAlias string) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (pfs *ProxyFileService) GetFileByContentId(contentId models.ContentId) (*fileTree.WeblensFileImpl, error) {
	panic("implement me")
}

func (pfs *ProxyFileService) GetFiles(ids []fileTree.FileId) ([]*fileTree.WeblensFileImpl, []fileTree.FileId, error) {
	if len(ids) == 0 {
		return []*fileTree.WeblensFileImpl{}, []fileTree.FileId{}, nil
	}

	r := NewCoreRequest(pfs.Core, "POST", "/files").WithBody(ids)

	type getFilesResponse struct {
		Files     []*fileTree.WeblensFileImpl `json:"files"`
		LostFiles []fileTree.FileId           `json:"lostFiles"`
	}

	res, err := CallHomeStruct[getFilesResponse](r)
	return res.Files, res.LostFiles, err
}

func (pfs *ProxyFileService) GetFileSafe(
	id fileTree.FileId, accessor *models.User, share *models.FileShare,
) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (pfs *ProxyFileService) GetFileTreeByName(treeName string) fileTree.FileTree {
	return nil
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
	files []*fileTree.WeblensFileImpl, destFolder *fileTree.WeblensFileImpl, treeName string, caster models.FileCaster,
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

func (pfs *ProxyFileService) DeleteFiles(files []*fileTree.WeblensFileImpl, treeName string, caster models.FileCaster) error {

	panic("implement me")
}

func (pfs *ProxyFileService) RestoreFiles(
	ids []fileTree.FileId, newParent *fileTree.WeblensFileImpl, restoreTime time.Time, caster models.FileCaster,
) error {

	panic("implement me")
}

func (pfs *ProxyFileService) RestoreHistory(lifetimes []*fileTree.Lifetime) error {

	panic("implement me")
}

func (pfs *ProxyFileService) ReadFile(f *fileTree.WeblensFileImpl) (io.ReadCloser, error) {
	resp, err := NewCoreRequest(pfs.Core, "GET", fmt.Sprintf("/file/content/%s", f.GetContentId())).Call()
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

func (pfs *ProxyFileService) GetJournalByTree(treeName string) fileTree.Journal {

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

func (pfs *ProxyFileService) NewBackupFile(lt *fileTree.Lifetime) (*fileTree.WeblensFileImpl, error) {
	panic("implement me")
}

func (pfs *ProxyFileService) GetFolderCover(folder *fileTree.WeblensFileImpl) (models.ContentId, error) {
	return "", nil
}

func (pfs *ProxyFileService) SetFolderCover(folderId fileTree.FileId, coverId models.ContentId) error {
	return nil
}

func (pfs *ProxyFileService) UserPathToFile(searchPath string, user *models.User) (*fileTree.WeblensFileImpl, error) {
	return nil, nil
}
