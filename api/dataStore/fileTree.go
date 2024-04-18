package dataStore

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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

var existingJournals []types.FileId

func fsTreeInsert(f, parent types.WeblensFile, c ...types.BroadcasterAgent) error {
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
	fileTree[f.Id()] = f
	parent.AddChild(f)

	if f.IsDir() {
		watcherAddDirectory(f)
		f.ReadDir()
	}

	if thisServer.IsCore() {
		if _, exist := slices.BinarySearch(existingJournals, f.Id()); !exist {
			JournalFileCreate(f)
		}
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

	parent.AddChild(f)

	if f.IsDir() {
		// if !f.Exists() {
		// 	util.Warning.Println("Creating directory that doesn't exist during insert to file tree")
		// 	err := f.CreateSelf()
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		watcherAddDirectory(f)
	}

	util.Each(c, func(c types.BroadcasterAgent) { c.PushFileCreate(f) })
	ResizeUp(f.GetParent(), c...)

	return nil
}

func fsTreeRemove(f types.WeblensFile, casters ...types.BroadcasterAgent) (err error) {
	// If the file does not already have an id, generating the id can lock the file tree,
	// so we must do that outside of the lock here to avoid deadlock
	if f.Id() == "" {
		return ErrNoFile
	}

	fsTreeLock.Lock()
	if fileTree[f.Id()] == nil {
		fsTreeLock.Unlock()
		util.Warning.Println("Tried to remove key not in FsTree", f.Id())
		return ErrNoFile
	}
	fsTreeLock.Unlock()

	f.GetParent().(*weblensFile).removeChild(f.Id())

	tasks := []types.Task{}

	f.RecursiveMap(func(file types.WeblensFile) {
		ts := file.GetTasks()
		tasks = append(tasks, ts...)
		util.Each(ts, func(t types.Task) { t.Cancel() })
		util.Each(file.GetShares(), func(s types.Share) { DeleteShare(s) })

		if file.IsDisplayable() {
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
		if file.Owner() != WEBLENS_ROOT_USER {
			JournalFileDelete(file)
		}
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

		// The file no longer has an id, so generating the id will lock the file tree,
		// we must do that outside of the lock here to avoid deadlock
		w.Id()

		fsTreeLock.Lock()
		fileTree[w.Id()] = w
		fsTreeLock.Unlock()

		w.GetParent().AddChild(w)

		if w.IsDisplayable() {
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

		JournalFileMove(preFile, w)
		util.Each(casters, func(c types.BroadcasterAgent) { c.PushFileMove(preFile, w) })
	})

	resizeMultiple(oldParent, f.GetParent(), casters...)

	return nil
}

func GetTreeSize() int {
	return len(fileTree)
}

func getAllFiles() []types.WeblensFile {
	return util.MapToSlicePure(fileTree)
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
