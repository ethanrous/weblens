package filetree

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/dataStore/history"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type fileTree struct {
	fMap       map[types.FileId]types.WeblensFile
	fsTreeLock *sync.Mutex
	journal    types.JournalService
	media      types.MediaRepo

	root         types.WeblensFile
	delDirectory types.WeblensFile
}

func boolPointer(b bool) *bool {
	return &b
}

func NewFileTree() types.FileTree {
	return &fileTree{
		fMap:       make(map[types.FileId]types.WeblensFile),
		fsTreeLock: &sync.Mutex{},
	}
}

func (ft *fileTree) NewFile(parent types.WeblensFile, filename string, isDir bool) types.WeblensFile {
	return &weblensFile{
		tree:     parent.GetTree(),
		parent:   parent.(*weblensFile),
		filename: filename,
		isDir:    boolPointer(isDir),
		owner:    parent.Owner(),

		tasksLock: &sync.Mutex{},
		childLock: &sync.Mutex{},
		children:  []*weblensFile{},

		size: -1,
	}
}

func (ft *fileTree) AddRoot(r types.WeblensFile) error {
	if !r.IsDir() {
		return ErrDirectoryRequired
	}
	self := r.(*weblensFile)
	// Root directory must be its own parent
	self.parent = self
	ft.addInternal(r.ID(), r)

	return nil
}

func (ft *fileTree) NewRoot(id types.FileId, filename, absPath string, owner types.User,
	parent types.WeblensFile) (types.WeblensFile, error) {

	f := &weblensFile{
		id:       id,
		filename: filename,
		owner:    owner,

		isDir:        boolPointer(true),
		absolutePath: absPath,

		childLock: &sync.Mutex{},
		children:  []*weblensFile{},
	}

	if parent == nil {
		f.parent = f
	} else {
		f.parent = parent.(*weblensFile)
	}

	ft.addInternal(f.ID(), f)
	return f, nil
}

func (ft *fileTree) addInternal(id types.FileId, f types.WeblensFile) {
	realF := f.(*weblensFile)
	if realF.tree == nil {
		realF.tree = ft
	}
	ft.fsTreeLock.Lock()
	defer ft.fsTreeLock.Unlock()

	// Do not use .ID() inside critical section, as it may need to use the locks
	ft.fMap[id] = f
}

func (ft *fileTree) deleteInternal(id types.FileId) {
	ft.fsTreeLock.Lock()
	defer ft.fsTreeLock.Unlock()
	delete(ft.fMap, id)
}

func (ft *fileTree) has(id types.FileId) bool {
	ft.fsTreeLock.Lock()
	defer ft.fsTreeLock.Unlock()
	_, ok := ft.fMap[id]
	return ok
}

func (ft *fileTree) Add(f, parent types.WeblensFile, c ...types.BroadcasterAgent) error {
	if ft.has(f.ID()) {
		return types.NewWeblensError(fmt.Sprintf("key collision on attempt to insert to filesystem tree: %s", f.ID()))
	}

	ft.addInternal(f.ID(), f)
	err := parent.AddChild(f)
	if err != nil {
		return err
	}

	if f.IsDir() {
		err = ft.journal.WatchFolder(f)
		if err != nil {
			return err
		}
	} else {
		err = ResizeUp(f, c...)
		if err != nil {
			util.ErrTrace(err)
		}
	}

	util.Each(c, func(c types.BroadcasterAgent) { c.PushFileCreate(f) })

	return nil
}

func (ft *fileTree) Del(f types.WeblensFile, casters ...types.BroadcasterAgent) (err error) {
	// If the file does not already have an id, generating the id can lock the file tree,
	// so we must do that outside of the lock here to avoid deadlock
	f.ID()

	realF := f.(*weblensFile)

	if !ft.has(realF.id) {
		util.Warning.Println("Tried to remove key not in FsTree", f.ID())
		return ErrNoFile
	}

	err = realF.parent.removeChild(f)
	if err != nil {
		return
	}

	var tasks []types.Task

	err = f.RecursiveMap(func(file types.WeblensFile) error {
		t := file.GetTask()
		if t != nil {
			tasks = append(tasks, t)
			t.Cancel()
		}
		util.Each(file.GetShares(), func(s types.Share) {
			err := dataStore.DeleteShare(s, ft)
			if err != nil {
				return
			}
		})

		if !file.IsDir() {
			contentId := file.GetContentId()
			m := ft.media.Get(contentId)
			if m != nil {
				m.RemoveFile(file)
			}

			// possibly bug: when a single delete event is deleting multiple of the same content id you get a collision
			// in the content folder

			backupF, _ := ft.delDirectory.GetChild(string(contentId))
			if contentId != "" && backupF == nil {
				backupF = ft.NewFile(ft.delDirectory, string(contentId), false)
				err = ft.Add(backupF, ft.delDirectory, casters...)
				if err != nil {
					return err
				}
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

		ft.deleteInternal(file.ID())

		return nil
	})

	if err != nil {
		return
	}

	for _, t := range tasks {
		t.Wait()
	}

	if f.IsDir() {
		err = os.RemoveAll(f.GetAbsPath())
		if err != nil {
			return
		}
	}

	// if len(casters) == 0 {
	// 	casters = append(casters, globalCaster)
	// }

	util.Each(casters, func(c types.BroadcasterAgent) { c.PushFileDelete(f) })

	return
}

func (ft *fileTree) Get(fileId types.FileId) types.WeblensFile {
	ft.fsTreeLock.Lock()
	defer ft.fsTreeLock.Unlock()
	return ft.fMap[fileId]
}

func (ft *fileTree) Move(f, newParent types.WeblensFile, newFilename string, overwrite bool, c ...types.BufferedBroadcasterAgent) error {
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
	err := f.RecursiveMap(func(w types.WeblensFile) error {
		t := w.GetTask()
		if t != nil {
			allTasks = append(allTasks, t)
			t.Cancel()
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, t := range allTasks {
		t.Wait()
	}

	oldAbsPath := f.GetAbsPath()
	oldParent := f.GetParent()

	// Point of no return //

	// Overwrite filename
	f.(*weblensFile).filename = newFilename

	// Disable casters because we need to wait to move the files before stat-ing them for the updates
	util.Each(c, func(c types.BufferedBroadcasterAgent) { c.DisableAutoFlush() })

	// Sync file tree with new move, including f and all of its children.
	err = f.RecursiveMap(func(w types.WeblensFile) error {
		preFile := w.Copy()

		realW := w.(*weblensFile)
		if f == w {
			realW.parent = newParent.(*weblensFile)
		}

		err := preFile.GetParent().(*weblensFile).removeChild(w)
		if err != nil {
			return err
		}

		// The file no longer has an id, so generating the id will lock the file tree,
		// we must do that outside the lock below to avoid deadlock
		w.ID()
		_, err = w.Size()
		if err != nil {
			return err
		}

		ft.deleteInternal(realW.id)

		realW.id = ""
		realW.absolutePath = filepath.Join(w.GetParent().GetAbsPath(), w.Filename())
		if realW.IsDir() {
			realW.absolutePath += "/"
		}

		// w.ID()
		ft.addInternal(realW.id, w)

		err = w.GetParent().AddChild(w)
		if err != nil {
			return err
		}

		if w.IsDisplayable(ft.media) {
			m := ft.media.Get(preFile.GetContentId())
			if m != nil {
				// Add new file first so the media doesn't get deleted if there is only 1 file
				m.AddFile(w)
				m.RemoveFile(preFile)
			}
		}

		for _, s := range w.GetShares() {
			s.SetItemId(string(w.GetContentId()))
			err := w.UpdateShare(s)
			if err != nil {
				return err
			}
		}

		util.Each(c, func(c types.BufferedBroadcasterAgent) { c.PushFileMove(preFile, w) })
		return nil
	})

	if err != nil {
		return err
	}

	err = os.Rename(oldAbsPath, newAbsPath)
	if err != nil {
		util.ErrTrace(err)
		return err
	}

	err = resizeMultiple(oldParent, f.GetParent(), util.SliceConvert[types.BroadcasterAgent](c)...)
	if err != nil {
		return err
	}

	util.Each(c, func(c types.BufferedBroadcasterAgent) { c.AutoFlushEnable() })

	return nil
}

// Size gets the number of files loaded into weblens.
// This does not lock the file tree, and therefore
// cannot be trusted to be microsecond accurate, but
// it's quite close
func (ft *fileTree) Size() int {
	return len(ft.fMap)
}

func (ft *fileTree) GetMediaRepo() types.MediaRepo {
	return ft.media
}

func (ft *fileTree) Touch(parentFolder types.WeblensFile, newFileName string, detach bool, owner types.User, c ...types.BroadcasterAgent) (types.WeblensFile, error) {
	f := ft.NewFile(parentFolder, newFileName, false).(*weblensFile)
	f.detached = detach
	e := ft.Get(f.ID())
	if e != nil || f.Exists() {
		return e, ErrFileAlreadyExists
	}

	err := f.CreateSelf()
	if err != nil {
		return f, err
	}

	// Detach creates the file on the real filesystem,
	// but does not add it to the tree or journal its creation
	if detach {
		return f, nil
	}

	err = ft.Add(f, parentFolder, c...)
	if err != nil {
		return f, err
	}

	if owner != nil {
		f.owner = owner
	}

	return f, nil
}

// MkDir creates a new dir as a child of parentFolder named newDirName. If the dir already exists,
// it will be returned along with a ErrDirAlreadyExists error.
func (ft *fileTree) MkDir(parentFolder types.WeblensFile, newDirName string, c ...types.BroadcasterAgent) (types.WeblensFile, error) {
	d := ft.NewFile(parentFolder, newDirName, true).(*weblensFile)

	if d.Exists() {
		existingFile := ft.Get(d.ID())

		if existingFile == nil {
			err := ft.Add(d, parentFolder, c...)
			if err != nil {
				return d, err
			}
			existingFile = d
		}

		return existingFile, ErrDirAlreadyExists
	}

	d.size = 0

	err := ft.Add(d, parentFolder, c...)
	if err != nil {
		return d, err
	}

	err = d.CreateSelf()
	if err != nil {
		return d, err
	}

	return d, nil
}

// AttachFile takes a detached file when it is ready to be inserted to the tree, and attaches it
func (ft *fileTree) AttachFile(f types.WeblensFile, c ...types.BroadcasterAgent) error {
	if ft.Get(f.ID()) != nil {
		return ErrFileAlreadyExists
	}

	tmpPath := filepath.Join("/tmp/", f.Filename())
	tmpFile, err := os.Open(tmpPath)
	if err != nil {
		return err
	}

	// Mask the file create event. Since we cannot insert
	// into the tree if the file is not present already,
	// we must move the file first, which would create an
	// insert event, which we wish to do manually below.
	// Masking prevents the automatic insert event.
	history.MaskEvent(dataStore.FileCreate, f.GetAbsPath())

	destFile, err := os.Create(f.GetAbsPath())
	if err != nil {
		return err
	}

	_, err = io.Copy(destFile, tmpFile)
	if err != nil {
		return err
	}

	err = ft.Add(f, f.GetParent(), c...)
	if err != nil {
		return err
	}

	return os.Remove(tmpPath)
}

func (ft *fileTree) GenerateFileId(absPath string) types.FileId {
	fileHash := types.FileId(util.GlobbyHash(8, FilepathFromAbs(absPath).ToPortable()))
	return fileHash
}

func (ft *fileTree) GetRoot() types.WeblensFile {
	if ft.root == nil {
		util.Error.Println("GetRoot called on fileTree with nil root")
	}
	return ft.root
}

func (ft *fileTree) SetRoot(root types.WeblensFile) {
	ft.root = root
}

func (ft *fileTree) GetJournal() types.JournalService {
	return ft.journal
}

func (ft *fileTree) SetJournal(j types.JournalService) {
	ft.journal = j
}

// Util

func resizeMultiple(old, new types.WeblensFile, c ...types.BroadcasterAgent) (err error) {
	// Check if either of the files are a parent of the other
	oldIsParent := strings.HasPrefix(old.GetAbsPath(), new.GetAbsPath())
	newIsParent := strings.HasPrefix(new.GetAbsPath(), old.GetAbsPath())

	if oldIsParent || !newIsParent {
		err = ResizeUp(old, c...)
		if err != nil {
			return
		}
	}

	if newIsParent || !oldIsParent {
		err = ResizeUp(new, c...)
		if err != nil {
			return
		}
	}

	return
}

func ResizeUp(f types.WeblensFile, c ...types.BroadcasterAgent) error {
	return f.BubbleMap(func(w types.WeblensFile) error {
		return w.(*weblensFile).LoadStat(c...)
	})
}

func ResizeDown(f types.WeblensFile, c ...types.BroadcasterAgent) error {
	return f.LeafMap(func(w types.WeblensFile) error {
		return w.(*weblensFile).LoadStat(c...)
	})
}

// Error

var ErrDirNotAllowed = types.NewWeblensError("attempted to perform action using a directory, where the action does not support directories")
var ErrDirectoryRequired = types.NewWeblensError("attempted to perform an action that requires a directory, but found regular file")
var ErrDirAlreadyExists = types.NewWeblensError("directory already exists in destination location")
var ErrFileAlreadyExists = types.NewWeblensError("file already exists in destination location")
var ErrNoFile = types.NewWeblensError("file does not exist")
var ErrIllegalFileMove = types.NewWeblensError("tried to perform illegal file move")
var ErrWriteOnReadOnly = types.NewWeblensError("tried to write to read-only file")
var ErrBadReadCount = types.NewWeblensError("did not read expected number of bytes from file")
var ErrNoFileAccess = types.NewWeblensError("user does not have access to file")
var ErrAlreadyWatching = types.NewWeblensError("trying to watch directory that is already being watched")
var ErrBadTask = types.NewWeblensError("did not get expected task id while trying to unlock file")
var ErrNoShare = types.NewWeblensError("could not find share")
