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
		if err := FsTreeInsert(homeDir, "0"); err != nil {
			panic(err)
		}
	}
}

func FsTreeInsert(f *WeblensFileDescriptor, parentId string) error {
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

	if !f.IsDir() {
		f.GetParent().BubbleMap(func(file *WeblensFileDescriptor) {
			size, err := f.Size()
			if err != nil {
				util.DisplayError(err)
				return
			}

			file.size = file.size + size
			caster.PushItemUpdate(file)
		})
	}

	caster.PushItemCreate(f)
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

	f.GetParent().removeChild(f.Id())

	f.RecursiveMap(func(file *WeblensFileDescriptor) {
		if file.IsDir() {
			fddb.deleteFolder(file)
		} else if file.IsDisplayable() {
			fddb.RemoveMediaByFile(file)
		}
		fsTreeLock.Lock()
		delete(fileTree, file.Id())
		fsTreeLock.Unlock()
		// util.Debug.Printf("Inner removed %s", file.String())
	})

	err := os.RemoveAll(f.absolutePath)
	if err != nil {
		util.DisplayError(err)
	} else {
		util.Info.Printf("Removed %s from tree", f.absolutePath)
	}

	f.GetParent().BubbleMap(func(file *WeblensFileDescriptor) {
		file.recompSize()
	})

	caster.PushItemDelete(f)
	tasker.ScanDirectory(f.parent, false, true)
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

func FsTreeMove(f, newParent *WeblensFileDescriptor, newFilename string, overwrite bool) error {
	if !newParent.IsDir() {
		return errors.New("cannot move file to a non-directory")
	}

	if (newFilename == "" || newFilename == f.Filename()) && newParent == f.GetParent() {
		util.Warning.Println("Exiting early from move without updates")
		return nil
	}

	if newFilename == "" {
		newFilename = f.Filename()
	}

	newAbsPath := filepath.Join(newParent.String(), newFilename)

	if !overwrite {
		if _, err := os.Stat(newAbsPath); err == nil {
			return errors.New("file already exists in destination location")
		}
	}

	if !f.Exists() || !newParent.Exists() {
		return fmt.Errorf("file or parent does not exist while trying to move")
	}
	fileSize, err := f.Size()
	if err != nil {
		return err
	}

	// Point of no return

	f.parent.removeChild(f.Id())
	err = os.Rename(f.String(), newAbsPath)
	if err != nil {
		// This really shouldn't happen, but anything is possible
		util.DisplayError(err)
		return err
	}

	f.GetParent().BubbleMap(func(w *WeblensFileDescriptor) {
		w.size -= fileSize
		caster.PushItemUpdate(w)
	})

	f.parent = newParent
	f.filename = newFilename
	f.RecursiveMap(func(w *WeblensFileDescriptor) {
		preFile := w.Copy()

		if w.IsDir() {
			fddb.deleteFolder(w)
		}

		fsTreeLock.Lock()
		delete(fileTree, w.Id())
		w.id = ""
		w.absolutePath = filepath.Join(w.GetParent().absolutePath, w.Filename())
		fileTree[w.Id()] = w
		fsTreeLock.Unlock()

		if w.IsDir() {
			fddb.writeFolder(w)
		} else {
			fddb.HandleMediaMove(preFile, w)
		}

		caster.PushItemMove(preFile, w)
	})

	f.GetParent().BubbleMap(func(w *WeblensFileDescriptor) {
		w.size += fileSize
		caster.PushItemUpdate(w)
	})

	f.GetParent().AddChild(f)

	return nil
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
