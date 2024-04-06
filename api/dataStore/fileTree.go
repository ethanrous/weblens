package dataStore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

var fileTree = map[types.FileId]types.WeblensFile{}
var fsTreeLock = &sync.Mutex{}

// Disables certain actions if we know we can batch them for later
// Mainly just for init of the fs, then is disabled for the rest of runtime
var safety bool = false

func FsTreeInsert(f, parent types.WeblensFile, c ...types.BroadcasterAgent) error {
	if f.Filename() == ".DS_Store" {
		return nil
	}

	if safety {
		return mainInsert(f, parent, c...)
	}
	return initInsert(f, parent)
}

func initInsert(f, parent types.WeblensFile) error {
	if f.Id() == "" {
		return fmt.Errorf("not inserting file with empty file id")
	}
	realF := f.(*weblensFile)

	fileTree[f.Id()] = f

	realF.parent = parent
	parent.AddChild(f)
	realF.readOnly = parent.IsReadOnly()

	realF.childLock = &sync.Mutex{}
	realF.children = map[types.FileId]types.WeblensFile{}
	realF.tasksLock = &sync.Mutex{}

	if f.IsDir() {
		f.ReadDir()
	}

	return nil
}

func mainInsert(f, parent types.WeblensFile, c ...types.BroadcasterAgent) error {
	if f.Id() == "" {
		return fmt.Errorf("not inserting file with empty file id")
	}

	fsTreeLock.Lock()
	if fileTree[f.Id()] != nil {
		fsTreeLock.Unlock()
		return fmt.Errorf("key collision on attempt to insert to filesystem tree: %s", f.Id()).(AlreadyExistsError)
	}
	fileTree[f.Id()] = f
	fsTreeLock.Unlock()

	realF := f.(*weblensFile)
	realF.parent = parent
	realF.childLock = &sync.Mutex{}
	realF.children = map[types.FileId]types.WeblensFile{}
	realF.tasksLock = &sync.Mutex{}
	parent.AddChild(f)

	if f.IsDir() {
		if !f.Exists() {
			util.Warning.Println("Creating directory that doesn't exist during insert to file tree")
			err := f.CreateSelf()
			if err != nil {
				return err
			}
		}
	}

	ResizeUp(f.GetParent(), c...)

	return nil
}

func FsTreeRemove(f types.WeblensFile, casters ...types.BroadcasterAgent) (err error) {
	fsTreeLock.Lock()
	if fileTree[f.Id()] == nil {
		fsTreeLock.Unlock()
		util.Warning.Println("Tried to remove key not in FsTree", f.Id())
		return
	}
	fsTreeLock.Unlock()

	f.GetParent().(*weblensFile).removeChild(f.Id())

	tasks := []types.Task{}

	f.RecursiveMap(func(file types.WeblensFile) {
		ts := file.GetTasks()
		tasks = append(tasks, ts...)
		util.Each(ts, func(t types.Task) { t.Cancel() })
		util.Each(file.GetShares(), func(s types.Share) { DeleteShare(s) })

		displayable, err := file.IsDisplayable()
		if err != nil && err != ErrDirNotAllowed && err != ErrNoMedia {
			return
		}

		if displayable {
			var m types.Media
			m, err = file.GetMedia()
			if err != nil {
				return
			}
			if m != nil {
				m.RemoveFile(file.Id())
				file.ClearMedia()
			}
		}

		fsTreeLock.Lock()
		delete(fileTree, file.Id())
		fsTreeLock.Unlock()
	})

	for _, t := range tasks {
		t.Wait()
	}

	err = os.RemoveAll(f.GetAbsPath())
	if err != nil {
		return
	}

	if len(casters) == 0 {
		casters = append(casters, globalCaster)
	}

	util.Each(casters, func(c types.BroadcasterAgent) { c.PushFileDelete(f) })

	return
}

func FsTreeGet(fileId types.FileId) (f types.WeblensFile) {
	fsTreeLock.Lock()
	f = fileTree[fileId]
	fsTreeLock.Unlock()

	return
}

func FsTreeMove(f, newParent types.WeblensFile, newFilename string, overwrite bool, casters ...types.BroadcasterAgent) error {
	if f.Owner() != newParent.Owner() {
		return ErrIllegalFileMove
	}
	if !newParent.IsDir() {
		return ErrDirectoryRequired
	}

	if (newFilename == "" || newFilename == f.Filename()) && newParent == f.GetParent() {
		util.Warning.Println("Exiting early from move without updates")
		return nil
	}

	if newFilename == "" {
		newFilename = f.Filename()
	}

	newAbsPath := filepath.Join(newParent.GetAbsPath(), newFilename)

	if !overwrite {
		// Check if the file at the destination exists already
		if _, err := os.Stat(newAbsPath); err == nil {
			return ErrFileAlreadyExists
		}
	}

	if !f.Exists() || !newParent.Exists() {
		return ErrNoFile
	}

	var allTasks []types.Task
	f.RecursiveMap(func(w types.WeblensFile) {
		for _, t := range w.GetTasks() {
			allTasks = append(allTasks, t)
			t.Cancel()
		}
	})

	for _, t := range allTasks {
		t.Wait()
	}

	// Point of no return
	err := os.Rename(f.GetAbsPath(), newAbsPath)
	if err != nil {
		// This really shouldn't happen, but anything is possible
		util.ErrTrace(err)
		return err
	}

	oldParent := f.GetParent()

	// Overwrite filename
	f.(*weblensFile).filename = newFilename

	if len(casters) == 0 {
		casters = append(casters, globalCaster)
	}

	// Sync file tree with new move, including f and all of its children.
	f.RecursiveMap(func(w types.WeblensFile) {
		preFile := w.Copy()

		if f == w {
			w.(*weblensFile).parent = newParent
		}

		preFile.GetParent().(*weblensFile).removeChild(w.Id())

		fsTreeLock.Lock()
		delete(fileTree, w.Id())
		fsTreeLock.Unlock()

		w.(*weblensFile).id = ""
		w.(*weblensFile).absolutePath = filepath.Join(w.GetParent().GetAbsPath(), w.Filename())

		fsTreeLock.Lock()
		fileTree[w.Id()] = w
		fsTreeLock.Unlock()

		w.GetParent().AddChild(w)

		if d, _ := w.IsDisplayable(); d {
			var m types.Media
			m, err = preFile.GetMedia()

			if err != nil && err != ErrNoMedia {
				util.ErrTrace(err)
				return
			} else if err != ErrNoMedia {
				// Add new file first so the media doesn't get deleted if there is only 1 file
				m.AddFile(w)
				m.RemoveFile(preFile.Id())
			}
		}

		for _, s := range w.GetShares() {
			s.SetContentId(w.Id().String())
			w.UpdateShare(s)
		}

		util.Each(casters, func(c types.BroadcasterAgent) { c.PushFileMove(preFile, w) })
	})

	resizeMultiple(oldParent, f.GetParent(), casters...)

	return nil
}

func GetTreeSize() int {
	return len(fileTree)
}

func ResizeUp(f types.WeblensFile, c ...types.BroadcasterAgent) {
	f.BubbleMap(func(w types.WeblensFile) {
		w.(*weblensFile).loadStat(c...)
	})
}

func ResizeDown(f types.WeblensFile, c ...types.BroadcasterAgent) {
	f.LeafMap(func(w types.WeblensFile) {
		w.(*weblensFile).loadStat(c...)
	})
}

func resizeMultiple(old, new types.WeblensFile, c ...types.BroadcasterAgent) {
	// Check if either of the files are a parent of the other
	oldIsParent := strings.HasPrefix(old.GetAbsPath(), new.GetAbsPath())
	newIsParent := strings.HasPrefix(new.GetAbsPath(), old.GetAbsPath())

	if oldIsParent || !(oldIsParent || newIsParent) {
		ResizeUp(old, c...)
	}

	if newIsParent || !(oldIsParent || newIsParent) {
		ResizeUp(new, c...)
	}

}
