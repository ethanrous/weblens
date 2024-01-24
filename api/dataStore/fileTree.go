package dataStore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/ethrousseau/weblens/api/util"
)

var fileTree = map[string]*WeblensFile{}
var fsTreeLock = &sync.Mutex{}

// Disables certain actions if we know we can batch them for later
// Mainly just for init of the fs, then is disabled for the rest of runtime
var safety bool = false

func FsTreeInsert(f, parent *WeblensFile) error {
	if f.Filename() == ".DS_Store" {
		return nil
	}

	if safety {
		return mainInsert(f, parent)
	}
	return initInsert(f, parent)
}

func initInsert(f, parent *WeblensFile) error {
	if f.Id() == "" {
		return fmt.Errorf("not inserting file with empty file id")
	}

	fileTree[f.id] = f

	f.parent = parent
	parent.AddChild(f)

	if f.IsDir() {
		f.ReadDir()
		if _, b := slices.BinarySearchFunc(initFolderIds, f.Id(), strings.Compare); !b {
			fddb.writeFolder(f)
		}
	}

	// util.Debug.Println("Init file", f.String())
	return nil
}

func mainInsert(f, parent *WeblensFile) error {
	if f.Id() == "" {
		return fmt.Errorf("not inserting file with empty file id")
	}

	fsTreeLock.Lock()
	if fileTree[f.id] != nil {
		fsTreeLock.Unlock()
		return fmt.Errorf("key collision on attempt to insert to filesystem tree: %s", f.id).(alreadyExists)
	}
	fileTree[f.id] = f
	fsTreeLock.Unlock()

	f.parent = parent
	parent.AddChild(f)

	if f.IsDir() {
		fddb.writeFolder(f)
		if !f.Exists() {
			util.Warning.Println("Creating directory that doesn't exist during insert to file tree")
			err := f.CreateSelf()
			if err != nil {
				return err
			}
		}
	}

	Resize(f.GetParent())

	// util.Debug.Println("Inserted file", f.String())
	return nil
}

func FsTreeRemove(f *WeblensFile) {
	fsTreeLock.Lock()
	if fileTree[f.Id()] == nil {
		fsTreeLock.Unlock()
		util.Warning.Println("Tried to remove key not in FsTree", f.Id())
		return
	}
	fsTreeLock.Unlock()

	f.GetParent().removeChild(f.Id())

	f.RecursiveMap(func(file *WeblensFile) {
		util.Each(f.GetTasks(), func(t Task) { t.Cancel() })

		displayable, err := file.IsDisplayable()
		if displayable && err == nil {
			fddb.RemoveMediaByFile(file)
		} else if errors.Is(err, ErrDirNotAllowed) {
			fddb.deleteFolder(file)
		}

		fsTreeLock.Lock()
		delete(fileTree, file.Id())
		fsTreeLock.Unlock()
	})

	err := os.RemoveAll(f.absolutePath)
	if err != nil {
		util.DisplayError(err)
	}

	caster.PushFileDelete(f)
	Resize(f.GetParent())

	util.Debug.Println("Removed file", f.String())
}

func FsTreeGet(fileId string) (f *WeblensFile) {
	fsTreeLock.Lock()
	f = fileTree[fileId]
	fsTreeLock.Unlock()

	if f == nil {
		folder := fddb.getFolderById(fileId)
		if folder.FolderId == "" {
			return
		}
		f = wfInitFromFolderData(folder)
		f.GetChildren()
	}
	return
}

func FsTreeMove(f, newParent *WeblensFile, newFilename string, overwrite bool) error {
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

	// Point of no return
	err := os.Rename(f.String(), newAbsPath)
	if err != nil {
		// This really shouldn't happen, but anything is possible
		util.DisplayError(err)
		return err
	}

	oldParent := f.GetParent()

	// Overwrite filename
	f.filename = newFilename

	// Sync database and file tree with new move, including f and all of its children.
	f.RecursiveMap(func(w *WeblensFile) {
		preFile := w.Copy()

		if f == w {
			w.parent = newParent
		}

		if w.IsDir() {
			fddb.deleteFolder(w)
		}

		preFile.GetParent().removeChild(w.Id())

		fsTreeLock.Lock()
		delete(fileTree, w.Id())
		fsTreeLock.Unlock()

		w.id = ""
		w.absolutePath = filepath.Join(w.GetParent().absolutePath, w.Filename())

		fsTreeLock.Lock()
		fileTree[w.Id()] = w
		fsTreeLock.Unlock()

		w.GetParent().AddChild(w)

		if w.IsDir() {
			fddb.writeFolder(w)
		} else {
			fddb.handleMediaMove(preFile, w)
		}

		caster.PushFileMove(preFile, w)
	})

	resizeMultiple(oldParent, f.GetParent())

	return nil
}

func wfInitFromFolderData(fd folderData) *WeblensFile {
	f := WeblensFile{}
	f.id = fd.FolderId
	f.absolutePath = GuaranteeAbsolutePath(fd.RelPath)

	f.childLock = &sync.Mutex{}
	f.tasksLock = &sync.Mutex{}
	f.children = map[string]*WeblensFile{}
	parent := FsTreeGet(fd.ParentFolderId)
	FsTreeInsert(&f, parent)

	return &f
}

func GetTreeSize() int {
	return len(fileTree)
}

func Resize(f *WeblensFile) {
	f.BubbleMap(func(w *WeblensFile) {
		w.recompSize()
	})
}

func resizeMultiple(old, new *WeblensFile) {
	oldIsParent := strings.HasPrefix(old.String(), new.String())
	newIsParent := strings.HasPrefix(new.String(), old.String())

	if oldIsParent || !(oldIsParent || newIsParent) {
		Resize(old)
	}

	if newIsParent || !(oldIsParent || newIsParent) {
		Resize(new)
	}

}
