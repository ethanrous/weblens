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

var existingBackups []backupFile

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

	fsTreeLock.Lock()
	if fileTree[f.Id()] != nil {
		fsTreeLock.Unlock()
		return ErrFileAlreadyExists
	}

	fileTree[f.Id()] = f
	fsTreeLock.Unlock()
	err := parent.AddChild(f)
	if err != nil {
		return err
	}

	if f.IsDir() {
		watcherAddDirectory(f)
		f.ReadDir()
	}

	if thisServer.IsCore() && f.Owner() != WEBLENS_ROOT_USER && f.Owner() != EXTERNAL_ROOT_USER {
		if bi, exist := slices.BinarySearchFunc(existingBackups, f.Id(), func(b backupFile, t types.FileId) int { return strings.Compare(string(b.FileId), t.String()) }); !exist {
			jeStream <- fileEvent{action: FileCreate, postFilePath: f.GetAbsPath()}
		} else {
			f.(*weblensFile).setContentId(existingBackups[bi].ContentId)
		}
	}

	return nil
}

func mainInsert(f, parent types.WeblensFile, c ...types.BroadcasterAgent) error {
	// Generate fileId outside of lock section to avoid deadlock
	f.Id()
	fsTreeLock.Lock()
	if fileTree[f.Id()] != nil {
		fsTreeLock.Unlock()
		return fmt.Errorf("key collision on attempt to insert to filesystem tree: %s", f.Id()).(AlreadyExistsError)
	}
	fileTree[f.Id()] = f
	fsTreeLock.Unlock()

	err := parent.AddChild(f)
	if err != nil {
		return err
	}

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
	f.Id()

	realF := f.(*weblensFile)

	fsTreeLock.Lock()
	if fileTree[realF.id] == nil {
		fsTreeLock.Unlock()
		util.Warning.Println("Tried to remove key not in FsTree", f.Id())
		return ErrNoFile
	}
	fsTreeLock.Unlock()

	err = realF.parent.removeChild(f)
	if err != nil {
		return
	}

	tasks := []types.Task{}

	err = f.RecursiveMap(func(file types.WeblensFile) error {
		ts := file.GetTasks()
		tasks = append(tasks, ts...)
		util.Each(ts, func(t types.Task) { t.Cancel() })
		util.Each(file.GetShares(), func(s types.Share) { DeleteShare(s) })

		if !file.IsDir() {

			//
			// possibly bug: when a single delete action is deleting multiple of the same content id you get a collision in the content folder
			//

			backupF, _ := contentRoot.GetChild(file.GetContentId())
			if backupF == nil {
				backupF = newWeblensFile(&contentRoot, file.GetContentId(), false)
				fsTreeInsert(backupF, &contentRoot, casters...)
				err = os.Rename(file.GetAbsPath(), backupF.GetAbsPath())
				if err != nil {
					return err
				}
			} else {
				err := os.Remove(file.GetAbsPath())
				if err != nil {
					return err
				}
			}
		}

		file.Id()

		fsTreeLock.Lock()
		delete(fileTree, file.Id())
		fsTreeLock.Unlock()

		return nil
	})

	if err != nil {
		return
	}

	for _, t := range tasks {
		t.Wait()
	}

	if !f.IsDir() {
		err = os.RemoveAll(f.GetAbsPath())
		if err != nil {
			return
		}
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

func FsTreeMove(f, newParent types.WeblensFile, newFilename string, overwrite bool, casters ...types.BufferedBroadcasterAgent) error {
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
	f.RecursiveMap(func(w types.WeblensFile) error {
		for _, t := range w.GetTasks() {
			allTasks = append(allTasks, t)
			t.Cancel()
		}

		return nil
	})

	for _, t := range allTasks {
		t.Wait()
	}

	oldAbsPath := f.GetAbsPath()
	oldParent := f.GetParent()

	// Point of no return //

	// Overwrite filename
	f.(*weblensFile).filename = newFilename

	// Disable casters because we need to wait to move the files before stat-ing them for the updates
	util.Each(casters, func(c types.BufferedBroadcasterAgent) { c.DisableAutoFlush() })

	// Sync file tree with new move, including f and all of its children.
	f.RecursiveMap(func(w types.WeblensFile) error {
		preFile := w.Copy()

		realW := w.(*weblensFile)
		if f == w {
			realW.parent = newParent.(*weblensFile)
		}

		preFile.GetParent().(*weblensFile).removeChild(w)

		// The file no longer has an id, so generating the id will lock the file tree,
		// we must do that outside the lock below to avoid deadlock
		w.Id()
		w.Size()

		fsTreeLock.Lock()
		delete(fileTree, realW.id)
		fsTreeLock.Unlock()

		realW.id = ""
		realW.absolutePath = filepath.Join(w.GetParent().GetAbsPath(), w.Filename())
		if realW.IsDir() {
			realW.absolutePath += "/"
		}

		w.Id()

		fsTreeLock.Lock()
		fileTree[realW.id] = w
		fsTreeLock.Unlock()

		err := w.GetParent().AddChild(w)
		if err != nil {
			return err
		}

		if w.IsDisplayable() {
			m, err := preFile.GetMedia()

			if err != nil && err != ErrNoMedia {
				return err
			} else if err != ErrNoMedia {
				// Add new file first so the media doesn't get deleted if there is only 1 file
				m.AddFile(w)
				m.RemoveFile(preFile.Id())
			}
		}

		for _, s := range w.GetShares() {
			s.SetContentId(w.Id().String())
			err := w.UpdateShare(s)
			if err != nil {
				return err
			}
		}

		util.Each(casters, func(c types.BufferedBroadcasterAgent) { c.PushFileMove(preFile, w) })
		return nil
	})

	err := os.Rename(oldAbsPath, newAbsPath)
	if err != nil {
		util.ErrTrace(err)
		return err
	}

	resizeMultiple(oldParent, f.GetParent(), util.SliceConvert[types.BroadcasterAgent](casters)...)

	util.Each(casters, func(c types.BufferedBroadcasterAgent) { c.AutoFlushEnable() })

	return nil
}

// Gets the number of files loaded into weblens.
// This does not lock the file tree, and therefore
// cannot be trusted to be microsecond accurate, but
// it's quite close
func GetTreeSize() int {
	return len(fileTree)
}

func ResizeUp(f types.WeblensFile, c ...types.BroadcasterAgent) {
	f.BubbleMap(func(w types.WeblensFile) error {
		w.(*weblensFile).loadStat(c...)
		return nil
	})
}

func ResizeDown(f types.WeblensFile, c ...types.BroadcasterAgent) {
	f.LeafMap(func(w types.WeblensFile) error {
		w.(*weblensFile).loadStat(c...)
		return nil
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
