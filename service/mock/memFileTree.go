package mock

import (
	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/werror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ fileTree.FileTree = (*MemFileTree)(nil)

type MemFileTree struct {
	journal fileTree.Journal
	fMap    map[fileTree.FileId]*fileTree.WeblensFileImpl
	root    *fileTree.WeblensFileImpl

	rootAlias string
}

func (ft *MemFileTree) ReplaceId(oldId, newId fileTree.FileId) error {
	// TODO implement me
	panic("implement me")
}

func (ft *MemFileTree) AbsToPortable(absPath string) (fileTree.WeblensFilepath, error) {

	panic("implement me")
}

func NewMemFileTree(rootAlias string) *MemFileTree {
	fs := &MemFileTree{
		rootAlias: rootAlias,
		fMap:      map[fileTree.FileId]*fileTree.WeblensFileImpl{},
	}
	root := fileTree.NewWeblensFile("ROOT", "media", nil, true)
	root.ReplaceRoot(rootAlias)
	root.SetMemOnly(true)

	fs.root = root
	fs.fMap["ROOT"] = root

	return fs
}

func (ft *MemFileTree) Get(id fileTree.FileId) *fileTree.WeblensFileImpl {
	return ft.fMap[id]
}

func (ft *MemFileTree) GetRoot() *fileTree.WeblensFileImpl {
	return ft.root
}

func (ft *MemFileTree) ReadDir(dir *fileTree.WeblensFileImpl) ([]*fileTree.WeblensFileImpl, error) {
	return []*fileTree.WeblensFileImpl{}, nil
}

func (ft *MemFileTree) Size() int {
	return len(ft.fMap)
}

func (ft *MemFileTree) GetJournal() fileTree.Journal {
	return ft.journal
}

func (ft *MemFileTree) SetJournal(journal fileTree.Journal) {
	ft.journal = journal
}

func (ft *MemFileTree) Add(file *fileTree.WeblensFileImpl) error {
	parent := ft.Get(file.GetParentId())
	if parent == nil {
		return werror.Errorf("Could not find parent")
	}

	err := parent.AddChild(file)
	if err != nil {
		return err
	}

	ft.fMap[file.ID()] = file

	return nil
}

func (ft *MemFileTree) Remove(id fileTree.FileId) ([]*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (ft *MemFileTree) Delete(id fileTree.FileId, event *fileTree.FileEvent) error {
	// TODO implement me
	panic("implement me")
}

func (ft *MemFileTree) Move(
	f, newParent *fileTree.WeblensFileImpl, newFilename string, overwrite bool, event *fileTree.FileEvent,
) ([]fileTree.MoveInfo, error) {

	panic("implement me")
}

func (ft *MemFileTree) Touch(
	parentFolder *fileTree.WeblensFileImpl, newFileName string, event *fileTree.FileEvent,
) (*fileTree.WeblensFileImpl, error) {

	panic("implement me")
}

func (ft *MemFileTree) MkDir(
	parentFolder *fileTree.WeblensFileImpl, newDirName string, event *fileTree.FileEvent,
) (*fileTree.WeblensFileImpl, error) {
	newDir := fileTree.NewWeblensFile(ft.GenerateFileId(), newDirName, parentFolder, true)
	err := ft.Add(newDir)
	if err != nil {
		return nil, err
	}

	return newDir, nil
}

func (ft *MemFileTree) SetRootAlias(alias string) error {
	panic("implement me")
}

func (ft *MemFileTree) PortableToAbs(portable fileTree.WeblensFilepath) (string, error) {

	panic("implement me")
}

func (ft *MemFileTree) GenerateFileId() fileTree.FileId {
	return fileTree.FileId(primitive.NewObjectID().Hex())
}

func (ft *MemFileTree) ResizeDown(anchor *fileTree.WeblensFileImpl, event *fileTree.FileEvent, updateCallback func(newFile *fileTree.WeblensFileImpl)) error {
	return nil
}

func (ft *MemFileTree) ResizeUp(anchor *fileTree.WeblensFileImpl, event *fileTree.FileEvent, updateCallback func(newFile *fileTree.WeblensFileImpl)) error {
	return nil
}
