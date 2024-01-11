package dataStore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ethrousseau/weblens/api/util"
)

var fileTree = map[string]*WeblensFileDescriptor{}
var fsTreeLock = &sync.Mutex{}


func FsInit() {
	fileTree["0"] = &mediaRoot
	fileTree["1"] = &tmpRoot
	fileTree["2"] = &trashRoot
	fileTree["3"] = &takeoutRoot

	homeDirs, err := getUserHomeDirectories()
	if err != nil {
		panic(err)
	}
	for _, homeDir := range homeDirs {
		homeDir.id = util.HashOfString(8, GuaranteeRelativePath(homeDir.absolutePath))
		homeDir.ParentFolderId = "0"
		if err := FsTreeInsert(homeDir, "0"); err != nil {
			panic(err)
		}
	}
}

func FsTreeInsert(f *WeblensFileDescriptor, parentId string) (error) {
	if f.Id() == "" || f.absolutePath == "" {
		return fmt.Errorf("not inserting file with empty file id: %s", f.absolutePath)
	}
	fsTreeLock.Lock()
	if fileTree[f.id] != nil {
		fsTreeLock.Unlock()
		return fmt.Errorf("key collision on attempt to insert to filesystem tree: %s", f.id)
	}
	fileTree[f.id] = f
	fsTreeLock.Unlock()
	f.parent = FsTreeGet(parentId)
	f.parent.AddChild(f)

	if f.IsDir() {
		fddb.writeFolder(f)
		f.GetChildren()
	}
	return nil
}

func FsTreeRemove(f *WeblensFileDescriptor) {
	fsTreeLock.Lock()
	if fileTree[f.Id()] == nil {
		fsTreeLock.Unlock()
		util.Warning.Println("Tried to remove key not in FsTree", f.Id())
		return
	}
	fsTreeLock.Unlock()

	subTree := []string{}
	f.RecursiveMap(func(file *WeblensFileDescriptor) {
		if file.IsDir() {
			fddb.deleteFolder(file)
		} else if file.IsDisplayable() {
			fddb.RemoveMediaByFile(file)
		}
		subTree = append(subTree, file.Id())
		util.Debug.Printf("Inner removed %s", file.String())
	})
	fsTreeLock.Lock()
	for _, c := range subTree {
		delete(fileTree, c)
	}
	fsTreeLock.Unlock()

	f.parent.removeChild(f.Id())
	util.Info.Printf("Removed %s from tree", f.absolutePath)
}

func FsTreeGet(wfdId string) (f *WeblensFileDescriptor) {
	fsTreeLock.Lock()
	f = fileTree[wfdId]
	fsTreeLock.Unlock()

	if f == nil {
		folder := fddb.getFolderById(wfdId)
		if folder.FolderId == "" {
			return
		}
		f = wfdInitFromFolderData(folder)
		f.GetChildren()
	}
	return
}

func FsTreeMove(file, newParent *WeblensFileDescriptor, newFilename string, overwrite bool) error {
	if !newParent.IsDir() {
		return errors.New("cannot move file to a non-directory")
	}
	if newFilename == "" {
		newFilename = file.Filename()
	}

	newAbsPath := filepath.Join(newParent.String(), newFilename)

	if !overwrite {
		if _, err := os.Stat(newAbsPath); err == nil{
			return errors.New("file already exists in destination location")
		}
	}

	if !file.Exists() || !newParent.Exists() {
		return fmt.Errorf("file or parent does not exist while trying to move")
	}

	oldParent := file.GetParent()
	file.parent.removeChild(file.Id())
	fsTreeLock.Lock()
	delete(fileTree, file.Id())
	fsTreeLock.Unlock()

	children := []*WeblensFileDescriptor{}

	if file.IsDir() {
		children = file.GetChildren()
	}

	oldParent.childLock.Lock()
	err := os.Rename(file.String(), newAbsPath)
	if err != nil {
		oldParent.childLock.Unlock()
		return err
	}
	oldParent.childLock.Unlock()

	file.id = ""
	file.absolutePath = newAbsPath
	file.filename = filepath.Base(file.absolutePath)
	file.Id()

	for _, c := range children {
		delete(fileTree, c.Id())
	}

	file.children = nil

	if !oldParent.IsDir() {
		util.Error.Println("Wha? why isnt the parent a dir?")
	}
	oldParent.recompSize()

	return FsTreeInsert(file, newParent.Id())
}

func wfdInitFromFolderData(fd folderData) *WeblensFileDescriptor {
	f := WeblensFileDescriptor{}
	f.id = fd.FolderId
	f.absolutePath = GuaranteeAbsolutePath(fd.RelPath)

	f.childLock = &sync.Mutex{}
	f.children = map[string]*WeblensFileDescriptor{}
	FsTreeInsert(&f, fd.ParentFolderId)

	return &f
}

func GetTreeSize() int {
	return len(fileTree)
}