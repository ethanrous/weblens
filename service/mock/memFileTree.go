package mock

import "github.com/ethrousseau/weblens/fileTree"

var _ fileTree.FileTree = (*MemFileTree)(nil)

type MemFileTree struct{}

func (ft *MemFileTree) Get(id fileTree.FileId) *fileTree.WeblensFileImpl {
	// TODO implement me
	panic("implement me")
}

func (ft *MemFileTree) GetRoot() *fileTree.WeblensFileImpl {
	// TODO implement me
	panic("implement me")
}

func (ft *MemFileTree) ReadDir(dir *fileTree.WeblensFileImpl) ([]*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (ft *MemFileTree) Size() int {
	// TODO implement me
	panic("implement me")
}

func (ft *MemFileTree) GetJournal() fileTree.JournalService {
	// TODO implement me
	panic("implement me")
}

func (ft *MemFileTree) SetJournal(service fileTree.JournalService) {
	// TODO implement me
	panic("implement me")
}

func (ft *MemFileTree) Add(file fileTree.WeblensFile) error {
	// TODO implement me
	panic("implement me")
}

func (ft *MemFileTree) Del(id fileTree.FileId, deleteEvent *fileTree.FileEvent) ([]*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (ft *MemFileTree) Move(
	f, newParent *fileTree.WeblensFileImpl, newFilename string, overwrite bool, event *fileTree.FileEvent,
) ([]fileTree.MoveInfo, error) {
	// TODO implement me
	panic("implement me")
}

func (ft *MemFileTree) Touch(
	parentFolder *fileTree.WeblensFileImpl, newFileName string, detach bool,
) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (ft *MemFileTree) MkDir(
	parentFolder *fileTree.WeblensFileImpl, newDirName string, event *fileTree.FileEvent,
) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (ft *MemFileTree) PortableToAbs(portable fileTree.WeblensFilepath) (string, error) {
	// TODO implement me
	panic("implement me")
}

func (ft *MemFileTree) GenerateFileId() fileTree.FileId {
	// TODO implement me
	panic("implement me")
}
